module Gitlab
  module Git
    class RemoteMirror
      attr_reader :repository, :ref_name, :ssh_auth, :only_branches_matching

      def initialize(repository, ref_name, ssh_auth:, only_branches_matching: [])
        @repository = repository
        @ref_name = ref_name
        @ssh_auth = ssh_auth
        @only_branches_matching = only_branches_matching
      end

      def update
        ssh_auth.setup do |env|
          updated_branches = changed_refs(local_branches, remote_branches)
          push_refs(default_branch_first(updated_branches.keys), env: env)
          delete_refs(local_branches, remote_branches, env: env)

          local_tags = refs_obj(repository.tags)
          remote_tags = refs_obj(repository.remote_tags(ref_name, env: env))

          updated_tags = changed_refs(local_tags, remote_tags)
          push_refs(updated_tags.keys, env: env)
          delete_refs(local_tags, remote_tags, env: env)
        end
      end

      private

      def local_branches
        @local_branches ||= refs_obj(
          repository.local_branches,
          only_refs_matching: only_branches_matching
        )
      end

      def remote_branches
        @remote_branches ||= refs_obj(
          repository.remote_branches(ref_name),
          only_refs_matching: only_branches_matching
        )
      end

      def refs_obj(refs, only_refs_matching: [])
        refs.each_with_object({}) do |ref, refs|
          next if only_refs_matching.present? && !only_refs_matching.include?(ref.name)

          refs[ref.name] = ref
        end
      end

      def changed_refs(local_refs, remote_refs)
        local_refs.select do |ref_name, ref|
          remote_ref = remote_refs[ref_name]

          remote_ref.nil? || ref.dereferenced_target != remote_ref.dereferenced_target
        end
      end

      # Put the default branch first so it works fine when remote mirror is empty.
      def default_branch_first(branches)
        return unless branches.present?

        default_branch, branches = branches.partition do |branch|
          repository.root_ref == branch
        end

        branches.unshift(*default_branch)
      end

      def push_refs(refs, env:)
        return unless refs.present?

        repository.push_remote_branches(ref_name, refs, env: env)
      end

      def delete_refs(local_refs, remote_refs, env:)
        refs = refs_to_delete(local_refs, remote_refs)

        return unless refs.present?

        repository.delete_remote_branches(ref_name, refs.keys, env: env)
      end

      def refs_to_delete(local_refs, remote_refs)
        default_branch_id = repository.commit.id

        remote_refs.select do |remote_ref_name, remote_ref|
          next false if local_refs[remote_ref_name] # skip if branch or tag exist in local repo

          remote_ref_id = remote_ref.dereferenced_target.try(:id)

          remote_ref_id && repository.ancestor?(remote_ref_id, default_branch_id)
        end
      end
    end
  end
end
