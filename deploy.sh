function cloneTargetRepo() {
    local tmpdir="$(mktemp -d)"
    local tmpRepo="$tmpdir/weberc2.github.io"
    local cloneOutput # https://superuser.com/a/1103711/116125
    cloneOutput="$(git clone "git@github.com:weberc2/weberc2.github.io.git" "$tmpRepo" 2>&1)"
    local retVal=$?
    if [[ $retVal -ne 0 ]]; then
        echo "$cloneOutput" 1>&2 # output to stderr
        exit $retVal
    fi
    echo $tmpRepo
}

function deploy() {
    local sourceDirectory="$1"
    local sourceVersion="$2"

    if [[ -z "$sourceDirectory" ]]; then
        echo "USAGE deploy SOURCE_DIRECTORY SOURCE_VERSION"
        exit 1
    fi

    if [[ -z "$sourceVersion" ]]; then
        echo "USAGE deploy SOURCE_DIRECTORY SOURCE_VERSION"
        exit 1
    fi

    local targetRepo="$(cloneTargetRepo)"
    local targetVersion="$(git -C "$targetRepo" log --pretty=%s -1)"
    if [[ "$sourceVersion" != "$targetVersion" ]]; then
        echo "Source version has diverged from the current deployed version; redeploying..."
        git -C "$targetRepo" rm -rf $targetRepo/* # clear target repo
        cp -r $sourceDirectory/* $targetRepo
        git -C "$targetRepo" add -A "$targetRepo"
        # Allowing empty because there are some categories of inputs that
        # produce the same output and we don't want to report a failure if this
        # is the case. I could try to detect this, but that introduces a lot of
        # risk because git and bash are both very error-prone; it's easier to
        # just add an empty commit and move on with life.
        git -C "$targetRepo" commit --allow-empty -m "$sourceVersion"
        git -C "$targetRepo" push origin master
    else
        echo "Nothing changed; no need to deploy"
    fi
}

deploy $@
