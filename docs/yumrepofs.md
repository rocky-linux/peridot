# yumrepofs

`yumrepofs` is a technology that creates and manages a virtual yum repo server. Packages and/or metadata do not exist on disk together but is instead stored in a blob storage (for example, S3).

## Goals

There are a few goals of `yumrepofs`. However, **the primary goal** is to be able to store and serve RPM packages and repositories without the need for a remote-backed storage solution like NFS or local server disks.

## Repository Serving

Most of `dnf`'s common commands has calls with a (mostly) simple lifecycle:

* `dnf` sends a request to a repository looking for `$BASEURL/repodata/repomd.xml`
* Server responds with the initial metadata and more information to find the rest (primary, filelists, modules, etc)
* `dnf` queries metadata listed in repomd.xml

The most important metadata file in this case is `primary.xml`. This is where `dnf` finds information about packages, and specifically where to fetch them from.

And this is where `yumrepofs` comes in: It populates `href` fields with a path like `Packages/$TASK_ID/$KEY_ID/$RPM_NAME`. This in turn signs an S3 URL and thus will redirect the user's dnf client appropriately.

## Updating the repository/repositories

After a build is completed in peridot, the following happens:

* Peridot runs `createrepo_c` on all individual RPMs - This creates necessary metadata and stores that information for use later
* When an update request is sent to `yumrepofsupdater` containing a specific build, `GetBuildArtifacts` is called - This fetches the RPMs *and* corresponding metadata
* `yumrepofsupdater` fetches the latest revision of that repository and architecture, unmarshalls it, and swaps the metadata from the `GetBuildArtifacts` call

Using this method, packages never need to coexist on disk to create a new repository in the traditional way, because the metadata could then be swapped into the correct place as needed.

As of this writing, in-repo artifacts that are no longer found (based on its NEVRA) will be removed. This is just like with `createrepo_c` on a normal, on-disk repository.

## Comps, groups, and you

Using Peridot Catalogs, comps and groups can be declared that is then applied to a revision of the target repository. [comps2peridot](https://github.com/rocky-linux/peridot-releng/tree/main/comps2peridot) can be used to generate comps in a format Peridot accepts, for example.

## SQLite/Hashed hrefs

`yumrepofs` has something called "hashed variants" that can be created on demand. Just like pungi's `hashed_directories` directive, this changes package href's to the format of `Packages/$INITIAL/$RPM_NAME`. For example: `Packages/b/bash-5.1.8-4.el9.x86_64.rpm`. If an end user decides to do a reposync from a hashed `yumrepofs` repository, they will get a sorted repository directory structure that has packages sorted by their first alphanumeric character, similar to the Rocky Linux repositories as well as the Fedora Project's repositories.

This particular variant also runs `sqliterepo_c` on metadata and stores the sqlite files in S3, just like any other `yumrepofs` artifact. This is particularly useful for cases where applications that manage repositories (or "content") rely on sqlite metadata in a repository.
