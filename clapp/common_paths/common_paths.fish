function common_paths
    # set --show argv
    __common_paths_validate_args $argv

    set -l curr (mktemp -t common_paths)
    set -l prev (mktemp -t common_paths)
    set -l common_paths_file (mktemp -t common_paths)
    __common_paths_ls_into $argv[1] $prev
    for i in (seq 2 (count $argv))
        __common_paths_ls_into $argv[$i] $curr
        comm -12 $prev $curr >$common_paths_file
        cat $common_paths_file >$prev
    end
    cat $common_paths_file
    rm $prev $curr $common_paths_file
end

function __common_paths_ls_into -a dir -a temp_file
    find $dir -type f | cut -d/ -f2- | sort >$temp_file
end

function __common_paths_usage
    echo "common_paths <dir1> <dir2> {...<dirN>}"
end

function __common_paths_validate_args
    # set --show argv
    if test (count $argv) -lt 2
        echo "Expected at least two arguments"
        __common_paths_usage
        return 1
    end
    set -lp invalid_paths
    for dir in $argv
        if test ! -d $dir
            set -a invalid_paths $dir
        end
    end
    if test (count $invalid_paths) -gt 0
        echo "ERROR:"
        for dir in $invalid_paths
            echo -e "\tinvalid directory: $dir"
        end
        return 1
    end
end
