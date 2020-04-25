# README

This is the source code for my blog.

# BUILD AND DEPLOY

1. Install [`builder`](https://github.com/weberc2/builder)
2. Build and deploy `builder run //:deploy_script`. This will:
    1. Install [`neon`](https://github.com/weberc2/neon), the static site
       generator, into the `builder` cache.
    2. Run `neon` to generate the static site HTML package in the `builder`
       cache.
    3. Build and run the deploy script for deploying that particular HTML
       package. Deploy script deploys to
       https://github.com/weberc2/weberc2.github.io.

**NOTE**: In order to deploy, the user must have an SSH identity that can push
to `git@github.com:weberc2/weberc2.github.io.git`. An SSH identity can be
provided to git via `export GIT_SSH_COMMAND="ssh -i $PRIVATE_KEY_FILE"`.