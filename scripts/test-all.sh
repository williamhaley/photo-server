#!/usr/bin/env bash

set -e

function usage()
{
cat <<EOF
"$(basename $0)" -d <path> -l <number>

Generate a text file with paths to random images found by recursively searching
the source directory (-d) limited to (-l) images. Then try various ways of
generating thumbnails with those sample images to compare performance.

Returns the path to the output file.

-d <path>        Directory path to source images.
-l <number>      Limit of how many files to include in the output.

EOF
}

directory=""
limit=100
size=800
iterations=5

while getopts ":d:l:" opt; do
    case ${opt} in
        d)
            directory=${OPTARG}
            ;;
        l)
            limit=${OPTARG}
            ;;
        \?)
            usage
            exit 0
            ;;
        *)
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND -1))

if [ -z "${directory}" ];
then
    usage
    exit 1
fi

(
    cd scripts
    go build ./gojpeg.go
    go build ./goepeg.go
    go build ./golibjpeg-turbo.go
    go build ./goepeg-thumb.go
)

output_file=`mktemp`
output_directory=`mktemp -d`

find "${directory}" \( -iname \*.jpg -o -iname \*.jpeg \) -type f | shuf -n ${limit} > "${output_file}"

echo "${output_directory}"

# TODO WFH Are all these properly rotating?

# This doesn't always work "thumbnails-exiftool.sh" ...
commands=("thumbnails-imagemagick.sh" "thumbnails-epeg.sh" "thumbnails-libvips.sh" "thumbnails-gojpeg.sh" "thumbnails-goepeg.sh" "thumbnails-golibjpeg-turbo.sh")

for command in ${commands[@]}; do
  for i in `seq 1 ${iterations}`;
    do
        start_time="$(date -u +%s)"
        count=0

        while IFS= read -r file;
        do
            ((count=count+1))
            ./scripts/${command} "${file}" "${output_directory}" $size
        done < <(cat "${output_file}")

        end_time="$(date -u +%s)"
        elapsed="$(($end_time-$start_time))"
        echo "${command}: processed ${count} thumbnails in ${elapsed} seconds"
    done
done

rm -rf "${output_directory}"
rm -f "${output_file}"
