# README

This is the source code for my blog.

# BUILD AND DEPLOY

1. Install [`neon`](https://github.com/weberc2/neon)
2. Remove the `./_output` directory if it exists, then run `neon build`
3. Run `bash deploy.sh ./_output "$commitMessage"` (note the quotation marks)

**NOTE**: In order to deploy, the user must have an SSH identity that can push
to `git@github.com:weberc2/weberc2.github.io.git`. An SSH identity can be
provided to git via `export GIT_SSH_COMMAND="ssh -i $PRIVATE_KEY_FILE"`.
