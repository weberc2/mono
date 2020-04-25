load("std/python", "py_source_library", "py_source_binary", "pypi")
load("std/command", "bash")
load("3rdParty/python", "typing_extensions", "black_check", black = "black")
load("nimbus/resources", nimbus_resources = "lib")
load("nimbus/core", nimbus_core = "lib")
load("nimbus/util", nimbus_util = "lib")

src = glob("setup.py", "src/**/*.py")

bin = py_source_binary(
    name = "bin",
    package_name = "blog_analytics",
    sources = src,
    dependencies = [ nimbus_resources, nimbus_core, nimbus_util ],
    entry_point = "main",
)

cfn_template = bash(
    name = "cfn_template",
    environment = {"GENERATE": bin},
    # We don't want to do `$GENERATE > $OUTPUT` in case there is an error.
    script = "$GENERATE > ./output.json && mv ./output.json $OUTPUT",
)

deploy_script = bash(
    name = "deploy_script",
    environment = {"TEMPLATE": cfn_template},
    script = """

echo "#!/bin/bash" >> ./output.sh
echo 'if [[ -z "$1" ]]; then echo "Missing STACK_NAME parameter"; exit 1; fi' >> ./output.sh
echo "aws cloudformation deploy --template-file "$TEMPLATE" --stack-name "\$1" --capabilities CAPABILITY_NAMED_IAM" >> ./output.sh
chmod +x ./output.sh
mv ./output.sh $OUTPUT
    """,
)

codegen_black = black_check(name = "codegen_black", sources = [src])