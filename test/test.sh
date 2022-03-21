#!/usr/bin/env bash

#
# - Copy the fixture into a temporary location
# - Modify it in place
# - Check to ensure the replaced file has 
#   the expected image references.
# 

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
fixtures_dir="${this_dir}/fixtures"
root_dir="${this_dir}/.."
bin_name="pin-deploy-imgs-in-csv"
pin_tool="${root_dir}/${bin_name}"
csv_fixture="clusterserviceversion.yaml"

# Check for requirements.
echo "Checking system for requirements."
requirements="jq skopeo mktemp stat grep"
for r in $requirements ; do
    which $r &>/dev/null \
        || { 
            echo "ERR Could not find requirement \"${r}\" in path." \
            && exit 1
        }
done

echo "Verifying the test binary is built and lives at a known path."
# Check that the pin-deploy-imgs-in-csv binary exists in the root directory.
stat "${root_dir}/${bin_name}" &>/dev/null \
    || {
        echo "ERR Could not find the \"${bin_name}\" in the root directory"
        echo "    Make sure to build the binary before running the tests!" \
        && exit 5
        }

# Copy the test csv into a temporary location so we don't
# clutter up our repository.
echo "Creating a temporary directory and copying test fixtures into it."
tempdir=$(mktemp -d)
cp -a "${fixtures_dir}/${csv_fixture}" "${tempdir}"

echo "Running the pin tool against the fixture."
# Run the tool without error. Do not discard this output!
"${pin_tool}" "${tempdir}/${csv_fixture}" \
    || {
        echo "ERR Could not run \"${pin_tool}\" successfully."
        echo "    The test will now exit."
        exit 10
       }   

failed_tests=0

echo "Test: the tool must not modify any images that are already referenced via digest."
existing_image="image: gcr.io/kubebuilder/kube-rbac-proxy@sha256:db06cc4c084dd0253134f156dddaaf53ef1c3fb3cc809e5d81711baa4029ea4c"
grep "${existing_image}" "${tempdir}/${csv_fixture}" &>/dev/null \
    || {
        echo "FAILED something happened to the image that was already pinned before running the tool."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }

echo "Test: the tool must resolve an images digest when it does not contain a tag at all."
latest_digest=$(skopeo inspect docker://quay.io/opdev/simple-demo-operator | jq -r .Digest)
latest_image="image: quay.io/opdev/simple-demo-operator@${latest_digest} # latest"

grep "${latest_image}" "${tempdir}/${csv_fixture}" &>/dev/null \
    || {
        echo "FAILED the image that did not have a tag in the file was not properly resolved as \"latest\"."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }

echo "Test: the tool must resolve an image that is referenced via tag to its digest, and inject the tag as a comment."
tagged_image="image: quay.io/opdev/simple-demo-operator@sha256:25ca9cb1f3dc7b8ce0aba4d3383cac20f5f5a1298fbbfde4a6adab7b1000cb0e # 0.0.3"
grep "${tagged_image}" "${tempdir}/${csv_fixture}" &>/dev/null \
    || {
        echo "FAILED the image that had a tag was not properly resolved."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }

# Evaluate if we failed any tests.
if [ "${failed_tests}" != "0" ]; then 
    echo "Some test failed! The temp directory was:"
    echo "   ${temp_dir}."
    echo "Exiting."
    exit 15
fi

echo "All tests passed!"
