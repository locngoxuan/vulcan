# exit when any command fails
set -e

# keep track of the last executed command
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
# echo an error message before exiting
trap 'echo "\"${last_command}\" command filed with exit code $?."' EXIT

mvn clean install -U -Drevision=1.0.0

plugin/deploy-to-jfrog --path=target/file --username=1 --password=2 

