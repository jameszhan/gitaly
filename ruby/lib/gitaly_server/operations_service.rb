module GitalyServer
  class OperationsService < Gitaly::OperationService::Service
    include Utils

    def user_create_tag(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)

          gitaly_user = get_param!(request, :user)
          user = Gitlab::Git::User.from_gitaly(gitaly_user)

          tag_name = get_param!(request, :tag_name)

          target_revision = get_param!(request, :target_revision)

          created_tag = repo.add_tag(tag_name, user: user, target: target_revision, message: request.message.presence)
          return Gitaly::UserCreateTagResponse.new unless created_tag

          rugged_commit = created_tag.dereferenced_target.rugged_commit
          commit = gitaly_commit_from_rugged(rugged_commit)
          tag = gitaly_tag_from_gitlab_tag(created_tag, commit)

          Gitaly::UserCreateTagResponse.new(tag: tag)
        rescue Gitlab::Git::Repository::InvalidRef => e
          raise GRPC::FailedPrecondition.new(e.message)
        rescue Gitlab::Git::Repository::TagExistsError
          return Gitaly::UserCreateTagResponse.new(exists: true)
        rescue Gitlab::Git::PreReceiveError => e
          return Gitaly::UserCreateTagResponse.new(pre_receive_error: set_utf8!(e.message))
        end
      end
    end

    def user_delete_tag(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)

          gitaly_user = get_param!(request, :user)
          user = Gitlab::Git::User.from_gitaly(gitaly_user)

          tag_name = get_param!(request, :tag_name)

          repo.rm_tag(tag_name, user: user)

          Gitaly::UserDeleteTagResponse.new
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserDeleteTagResponse.new(pre_receive_error: set_utf8!(e.message))
        rescue Gitlab::Git::Repository::InvalidRef, Gitlab::Git::CommitError => e
          raise GRPC::FailedPrecondition.new(e.message)
        end
      end
    end

    def user_create_branch(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          target = get_param!(request, :start_point)
          gitaly_user = get_param!(request, :user)

          branch_name = request.branch_name
          user = Gitlab::Git::User.from_gitaly(gitaly_user)
          created_branch = repo.add_branch(branch_name, user: user, target: target)
          return Gitaly::UserCreateBranchResponse.new unless created_branch

          rugged_commit = created_branch.dereferenced_target.rugged_commit
          commit = gitaly_commit_from_rugged(rugged_commit)
          branch = Gitaly::Branch.new(name: branch_name, target_commit: commit)
          Gitaly::UserCreateBranchResponse.new(branch: branch)
        rescue Gitlab::Git::Repository::InvalidRef, Gitlab::Git::CommitError => ex
          raise GRPC::FailedPrecondition.new(ex.message)
        rescue Gitlab::Git::PreReceiveError => ex
          return Gitaly::UserCreateBranchResponse.new(pre_receive_error: set_utf8!(ex.message))
        end
      end
    end

    def user_update_branch(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          branch_name = get_param!(request, :branch_name)
          newrev = get_param!(request, :newrev)
          oldrev = get_param!(request, :oldrev)
          gitaly_user = get_param!(request, :user)

          user = Gitlab::Git::User.from_gitaly(gitaly_user)
          repo.update_branch(branch_name, user: user, newrev: newrev, oldrev: oldrev)

          Gitaly::UserUpdateBranchResponse.new
        rescue Gitlab::Git::Repository::InvalidRef, Gitlab::Git::CommitError => ex
          raise GRPC::FailedPrecondition.new(ex.message)
        rescue Gitlab::Git::PreReceiveError => ex
          return Gitaly::UserUpdateBranchResponse.new(pre_receive_error: set_utf8!(ex.message))
        end
      end
    end

    def user_delete_branch(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          user = Gitlab::Git::User.from_gitaly(request.user)

          repo.rm_branch(request.branch_name, user: user)

          Gitaly::UserDeleteBranchResponse.new
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserDeleteBranchResponse.new(pre_receive_error: set_utf8!(e.message))
        rescue Gitlab::Git::Repository::InvalidRef, Gitlab::Git::CommitError => e
          raise GRPC::FailedPrecondition.new(e.message)
        end
      end
    end

    def user_merge_branch(session, call)
      Enumerator.new do |y|
        bridge_exceptions do
          first_request = session.next

          repository = Gitlab::Git::Repository.from_gitaly(first_request.repository, call)
          user = Gitlab::Git::User.from_gitaly(first_request.user)
          source_sha = first_request.commit_id.dup
          target_branch = first_request.branch.dup
          message = first_request.message.dup

          begin
            result = repository.merge(user, source_sha, target_branch, message) do |commit_id|
              y << Gitaly::UserMergeBranchResponse.new(commit_id: commit_id)

              second_request = session.next
              raise GRPC::FailedPrecondition.new('merge aborted by client') unless second_request.apply
            end

            y << Gitaly::UserMergeBranchResponse.new(branch_update: branch_update_result(result))
          rescue Gitlab::Git::PreReceiveError => e
            y << Gitaly::UserMergeBranchResponse.new(pre_receive_error: set_utf8!(e.message))
          end
        end
      end
    end

    def user_ff_branch(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          user = Gitlab::Git::User.from_gitaly(request.user)

          result = repo.ff_merge(user, request.commit_id, request.branch)
          branch_update = branch_update_result(result)

          Gitaly::UserFFBranchResponse.new(branch_update: branch_update)
        rescue Gitlab::Git::CommitError => e
          raise GRPC::FailedPrecondition.new(e.to_s)
        rescue ArgumentError => e
          raise GRPC::InvalidArgument.new(e.to_s)
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserFFBranchResponse.new(pre_receive_error: set_utf8!(e.message))
        end
      end
    end

    def user_cherry_pick(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          user = Gitlab::Git::User.from_gitaly(request.user)
          commit = Gitlab::Git::Commit.new(repo, request.commit)
          start_repository = Gitlab::Git::GitalyRemoteRepository.new(request.start_repository || request.repository, call)

          result = repo.cherry_pick(
            user: user,
            commit: commit,
            branch_name: request.branch_name,
            message: request.message.dup,
            start_branch_name: request.start_branch_name.presence,
            start_repository: start_repository
          )

          branch_update = branch_update_result(result)
          Gitaly::UserCherryPickResponse.new(branch_update: branch_update)
        rescue Gitlab::Git::Repository::CreateTreeError => e
          Gitaly::UserCherryPickResponse.new(create_tree_error: set_utf8!(e.message))
        rescue Gitlab::Git::CommitError => e
          Gitaly::UserCherryPickResponse.new(commit_error: set_utf8!(e.message))
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserCherryPickResponse.new(pre_receive_error: set_utf8!(e.message))
        end
      end
    end

    def user_revert(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          user = Gitlab::Git::User.from_gitaly(request.user)
          commit = Gitlab::Git::Commit.new(repo, request.commit)
          start_repository = Gitlab::Git::GitalyRemoteRepository.new(request.start_repository || request.repository, call)

          result = repo.revert(
            user: user,
            commit: commit,
            branch_name: request.branch_name,
            message: request.message.dup,
            start_branch_name: request.start_branch_name.presence,
            start_repository: start_repository
          )

          branch_update = branch_update_result(result)
          Gitaly::UserRevertResponse.new(branch_update: branch_update)
        rescue Gitlab::Git::Repository::CreateTreeError => e
          Gitaly::UserRevertResponse.new(create_tree_error: set_utf8!(e.message))
        rescue Gitlab::Git::CommitError => e
          Gitaly::UserRevertResponse.new(commit_error: set_utf8!(e.message))
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserRevertResponse.new(pre_receive_error: set_utf8!(e.message))
        end
      end
    end

    def user_rebase(request, call)
      bridge_exceptions do
        begin
          repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
          user = Gitlab::Git::User.from_gitaly(request.user)
          remote_repository = Gitlab::Git::GitalyRemoteRepository.new(request.remote_repository, call)
          rebase_sha = repo.rebase(user, request.rebase_id,
                                   branch: request.branch,
                                   branch_sha: request.branch_sha,
                                   remote_repository: remote_repository,
                                   remote_branch: request.remote_branch)

          Gitaly::UserRebaseResponse.new(rebase_sha: rebase_sha)
        rescue Gitlab::Git::PreReceiveError => e
          return Gitaly::UserRebaseResponse.new(pre_receive_error: set_utf8!(e.message))
        rescue Gitlab::Git::Repository::GitError => e
          return Gitaly::UserRebaseResponse.new(git_error: set_utf8!(e.message))
        rescue Gitlab::Git::CommitError => e
          raise GRPC::FailedPrecondition.new(e.message)
        end
      end
    end

    def user_commit_files(call)
      bridge_exceptions do
        begin
          actions = []
          request_enum = call.each_remote_read
          header = request_enum.next.header

          loop do
            action = request_enum.next.action

            if action.header
              actions << commit_files_action_from_gitaly_request(action.header)
            else
              actions.last[:content] << action.content
            end
          end

          repo = Gitlab::Git::Repository.from_gitaly(header.repository, call)
          user = Gitlab::Git::User.from_gitaly(header.user)
          opts = commit_files_opts(call, header, actions)

          branch_update = branch_update_result(repo.multi_action(user, opts))

          Gitaly::UserCommitFilesResponse.new(branch_update: branch_update)
        rescue Gitlab::Git::Index::IndexError => e
          Gitaly::UserCommitFilesResponse.new(index_error: set_utf8!(e.message))
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserCommitFilesResponse.new(pre_receive_error: set_utf8!(e.message))
        rescue Gitlab::Git::CommitError => e
          raise GRPC::FailedPrecondition.new(e.message)
        end
      end
    end

    def user_squash(request, call)
      bridge_exceptions do
        repo = Gitlab::Git::Repository.from_gitaly(request.repository, call)
        user = Gitlab::Git::User.from_gitaly(request.user)
        author = Gitlab::Git::User.from_gitaly(request.author)

        begin
          squash_sha = repo.squash(user, request.squash_id,
                                   branch: request.branch,
                                   start_sha: request.start_sha,
                                   end_sha: request.end_sha,
                                   author: author,
                                   message: request.commit_message)

          Gitaly::UserSquashResponse.new(squash_sha: squash_sha)
        rescue Gitlab::Git::Repository::GitError => e
          Gitaly::UserSquashResponse.new(git_error: set_utf8!(e.message))
        end
      end
    end

    def user_apply_patch(call)
      bridge_exceptions do
        stream = call.each_remote_read
        first_request = stream.next

        header = first_request.header
        user = Gitlab::Git::User.from_gitaly(header.user)
        target_branch = header.target_branch
        patches = stream.lazy.map(&:patches)

        branch_update = Gitlab::Git::Repository.from_gitaly_with_block(header.repository, call) do |repo|
          begin
            Gitlab::Git::CommitPatches.new(user, repo, target_branch, patches).commit
          rescue Gitlab::Git::PatchError => e
            raise GRPC::FailedPrecondition.new(e.message)
          end
        end

        Gitaly::UserApplyPatchResponse.new(branch_update: branch_update_result(branch_update))
      end
    end

    def user_update_submodule(request, call)
      bridge_exceptions do
        user = Gitlab::Git::User.from_gitaly(request.user)

        begin
          branch_update = Gitlab::Git::Repository.from_gitaly_with_block(request.repository, call) do |repo|
            begin
              Gitlab::Git::Submodule
                .new(user, repo, request.submodule, request.branch)
                .update(request.commit_sha, request.commit_message.dup)
            rescue ArgumentError => e
              raise GRPC::InvalidArgument.new(e.to_s)
            end
          end

          Gitaly::UserUpdateSubmoduleResponse.new(branch_update: branch_update_result(branch_update))
        rescue Gitlab::Git::CommitError => e
          Gitaly::UserUpdateSubmoduleResponse.new(commit_error: set_utf8!(e.message))
        rescue Gitlab::Git::PreReceiveError => e
          Gitaly::UserUpdateSubmoduleResponse.new(pre_receive_error: set_utf8!(e.message))
        end
      end
    end

    private

    def commit_files_opts(call, header, actions)
      opts = {
        branch_name: header.branch_name,
        message: header.commit_message.b,
        actions: actions
      }

      opts[:start_repository] = Gitlab::Git::GitalyRemoteRepository.new(header.start_repository, call) if header.start_repository

      optional_fields = {
        start_branch_name: 'start_branch_name',
        author_name: 'commit_author_name',
        author_email: 'commit_author_email'
      }.transform_values { |v| header[v].presence }

      opts.merge(optional_fields)
    end

    def commit_files_action_from_gitaly_request(header)
      {
        action: header.action.downcase,
        # Forcing the encoding to UTF-8 here is unusual. But these paths get
        # compared with Rugged::Index entries, which are also force-encoded to
        # UTF-8. See
        # https://github.com/libgit2/rugged/blob/f8172c2a177a6795553f38f01248daff923f4264/ext/rugged/rugged_index.c#L514
        file_path: set_utf8!(header.file_path),
        previous_path: set_utf8!(header.previous_path),
        encoding: header.base64_content ? 'base64' : '',
        content: '',
        execute_filemode: header.execute_filemode
      }
    end

    def branch_update_result(gitlab_update_result)
      return if gitlab_update_result.nil?

      Gitaly::OperationBranchUpdate.new(
        commit_id: gitlab_update_result.newrev,
        repo_created: gitlab_update_result.repo_created,
        branch_created: gitlab_update_result.branch_created
      )
    end

    def get_param!(request, name)
      value = request[name.to_s]

      return value if value.present?

      field_name = name.to_s.tr('_', ' ')
      raise GRPC::InvalidArgument.new("empty #{field_name}")
    end
  end
end
