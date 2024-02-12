# How to release

# Minor


# Patch

1. Ensure all needed PRs were backported to the current release branch, and all backport PRs were merged.

2. Export the release you want to make:

    ```sh
    export REL_VER=v0.1.1
    export CUR_VER="$(echo $REL_VER | awk -F. -v OFS=. '{$NF -= 1 ; print}')"

    echo "Releasing $CUR_VER -> $REL_VER"
    ```
3. Checkout the current release branch:

   ```sh
   git switch "release/${REL_VER%.*}"
   ```

4. Create a new temporary branch for the relese:

    ```sh
    git switch -c "tmp/$REL_VER"
    git push
    ```

5. Trigger the release workflow

    ```sh
    gh workflow run release.yml --ref $(git rev-parse --abbrev-ref HEAD) -f kind=patch -f version="$REL_VER"
    ```
6. Review the release notes, test the binary artifact.

7. Publish the GitHub release.