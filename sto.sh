#!/usr/bin/env bash

set -e

state_path=~/.local/state/sto/
state_file=${state_path}current_profile

function set_current_profile() { 
    if [ ! -d $state_path ]; then
        mkdir -p $state_path
    fi
    echo $1 > $state_file
    return 0
}

function load_store() {
    if [ ! -f $state_file ]; then
        echo "No store set"
        exit 1
    fi

    current_store_root=$(cat $state_file)
    current_store_file=$current_store_root/.sto

    if [ ! -f $current_store_file ]; then
        echo "No store found at $current_store_root"
        exit 1
    fi
}

function apply_link() {
    source_file=$1
    destination_file=$2

    escaped_destination_file=${destination_file/#\~/$HOME}

    if [ -e $escaped_destination_file ]; then
        if [ -L $escaped_destination_file ]; then
            if [ "$(readlink $escaped_destination_file)" != "$current_store_root/$source_file" ]; then
                echo "Conflicting symlink at $escaped_destination_file" >&2
                return 1
            fi
            echo "[LINKED] $source_file -> $destination_file"
            return 0
        fi
        echo "Conflicting file at $escaped_destination_file" >&2
    fi

    if [ ! -d $(dirname $escaped_destination_file) ]; then
        mkdir -p $(dirname $escaped_destination_file)
    fi

    ln -snf $current_store_root/$source_file $escaped_destination_file

    echo "[LINKED] $source_file -> $destination_file"
}

function remove_link() {
    source_file=$1
    destination_file=$2

    escaped_destination_file=${destination_file/#\~/$HOME}

    if [ ! -L $escaped_destination_file ]; then
        echo "[UNLINKED] $source_file -> $destination_file"
        return 0
    fi

    if [ "$(readlink $escaped_destination_file)" != "$current_store_root/$source_file" ]; then
        echo "[UNLINKED] $source_file -> $destination_file"
        return 0
    fi

    rm $escaped_destination_file

    echo "[UNLINKED] $source_file -> $destination_file"
}

function run_init() {
    if [ $# -lt 1 ]; then
        echo "Usage sto init [path]" >&2
        exit 1
    fi

    store_root=$(realpath $1)

    if [ ! -d $store_root ]; then
        echo "$store_root is not a directory" >&2
        exit 1
    fi

    store_file_path=${store_root}/.sto

    if [ -f $store_file_path ]; then
        echo "Store already exists at $store_root" >&2
        exit 1
    fi

    touch $store_file_path

    echo "Created new store"

    if [ ! -f $state_file ]; then
        set_current_profile $store_root
    fi
}

function run_list() {
    load_store

    while read current_line; do
        source_file=$(echo $current_line | cut -f1 -d=)
        destination_file=$(echo $current_line | cut -f2 -d=)

        if [ ! -e $current_store_root/$source_file ]; then
            echo -e "[BROKEN] $source_file -> $destination_file"
            continue
        fi

        escaped_destination_file=${destination_file/#\~/$HOME}

        if [ ! -L $escaped_destination_file ]; then
            echo -e "[UNLINKED] $source_file -> $destination_file"
            continue
        fi

        if [ "$(readlink $escaped_destination_file)" != "$current_store_root/$source_file" ]; then
            echo -e "[CONFLICT] $source_file -> $destination_file"
            continue
        fi

        echo "[LINKED] $source_file -> $destination_file"
    done < $current_store_file
}

function run_link() {
    load_store

    while read current_line; do
        source_file=$(echo $current_line | cut -f1 -d=)
        destination_file=$(echo $current_line | cut -f2 -d=)

        if [ "$source_file" != "$1" ]; then
            continue
        fi

        apply_link $source_file $destination_file

        return 0
    done < $current_store_file

    echo "Entry $1 not found" >&2
    return 1
}

function run_unlink() {
    load_store

    while read current_line; do
        source_file=$(echo $current_line | cut -f1 -d=)
        destination_file=$(echo $current_line | cut -f2 -d=)

        if [ "$source_file" != "$1" ]; then
            continue
        fi

        remove_link $source_file $destination_file

        return 0
    done < $current_store_file

    echo "Entry $1 not found" >&2

    return 1
}

function run_apply() {
    if [ $# -lt 1 ]; then
        echo "Usage sto apply [path]" >&2
        exit 1
    fi

    store_root=$(realpath $1)
    store_file_path=${store_root}/.sto

    if [ ! -f $store_file_path ]; then
        echo "Store at $store_root not found" >&2
        exit 1
    fi

    if [ -f $state_file ]; then
        load_store

        while read current_line; do
            source_file=$(echo $current_line | cut -f1 -d=)
            destination_file=$(echo $current_line | cut -f2 -d=)

            remove_link $source_file $destination_file
        done < $current_store_file
    fi

    set_current_profile $store_root
    load_store

    while read current_line; do
        source_file=$(echo $current_line | cut -f1 -d=)
        destination_file=$(echo $current_line | cut -f2 -d=)

        apply_link $source_file $destination_file
    done < $current_store_file
}

case $1 in
    init)
        run_init ${@:2}
        ;;
    list)
        run_list ${@:2}
        ;;
    link)
        run_link ${@:2}
        ;;
    unlink)
        run_unlink ${@:2}
        ;;
    apply)
        run_apply ${@:2}
        ;;
    *)
        echo "$1 is not a valid command"
        exit 1
        ;;
esac
