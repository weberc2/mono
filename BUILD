load("std/golang", "go_module")
load("std/git", "git_clone")
load("std/bash", "bash")
load("neon", "neon")
load("system/stdenv", "runInStdenv")

def _blog(name, site_root="https://weberc2.github.io"):
    return runInStdenv(
        name = name,
        environment = {
            "SOURCES": glob("neon.yaml", "posts/**", "themes/**"),
            "NEON": neon,
            "SITE_ROOT": site_root,
        },
        script = """
        set -eo pipefail
        cp -r $SOURCES ./sources
        cd ./sources
        sed -i.bak 's|https://weberc2.github.io|{}|g' ./neon.yaml
        $NEON build
        mv _output $OUTPUT
        """.format(site_root),
    )

def _dockerTar(name, sources):
    """Builds a .tar.gz that's ready to be unpacked and `docker build`-ed.

    Contents:
    - site/ # static site root
    - Dockerfile
    - Caddyfile
    """
    return runInStdenv(
        name = name,
        environment = {
            "SOURCES": sources,
            "DOCKERDIR": glob("docker/**"),
        },
        script = """
        cp -r "$SOURCES" ./site
        cp "$DOCKERDIR/docker/Dockerfile" .
        cp "$DOCKERDIR/docker/Caddyfile" .
        tar -czvf output.tar.gz *
        mv output.tar.gz $OUTPUT
        """,
    )

blog = _blog("blog")

rpiTar = _dockerTar(
    "rpi-tar",
    _blog("rpi", site_root="http://blog.home/blog/"),
)

deploy_script = runInStdenv(
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
