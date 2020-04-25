load("std/golang", "go_module")
load("std/git", "git_clone")
load("std/command", "bash")
load("neon", "neon") 

blog = bash(
    name = "blog",
    environment = {
        "SOURCES": glob("neon.yaml", "posts/**", "themes/**"),
        "NEON": neon,
    },
    script = 'cd $SOURCES && $NEON build && mv _output $OUTPUT',
)

# NOTE this depends on `md5sum` and `awk` system utilities.
deploy_script = bash(
    name = "deploy_script",
    environment = {
        "SOURCE_DIRECTORY": blog,
        "SOURCES": glob("./deploy.sh"),
    },
    script = "\n".join([
        # Because the $SOURCE_DIRECTORY is the same for every build of a given
        # version of the blog source, we can treat it as an approximate
        # identifier for the blog source. Note that this only holds when the
        # cache path is held constant--builds of the same blog source version
        # in different systems will produce different $SOURCE_DIRECTORY values;
        # however, this will be a rare case for this application.
        """SOURCE_VERSION="$(echo "$SOURCE_DIRECTORY" | md5sum | awk '{ print $1 }')" """,
        'tmpfile=$(mktemp)',
        'echo "bash "$SOURCES/deploy.sh" "$SOURCE_DIRECTORY" "$SOURCE_VERSION"" > "$tmpfile"',
        'chmod +x $tmpfile',
        'mv $tmpfile $OUTPUT',
    ]),
)