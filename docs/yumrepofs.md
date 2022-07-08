# yumrepofs

Yumrepofs is a virtual Yum repo server. Packages or metadata do not co-exist on disk, but is rather stored in blob storage (ex. S3)

The goal is to be able to store RPMs and serve repositories without the need for NFS.

## Serving

The lifecycle of a dnf call is as follows:

* dnf sends a request to {baseurl}/repodata/repomd.xml
* Server responds with metadata on where to find further metadata (ex. primary, filelists and other)
* dnf queries metadata listed in repomd.xml

The most important file in this case is definitely primary.xml. This is where dnf finds information about packages, and specifically where to fetch them from.

Yumrepofs populates `href` fields with a path like this `Packages/{TASK_ID}/{KEY_ID}/{RPM_NAME}`. This in turn signes a S3 url and redirects the end user.

## Updating the repo

Let's start from the build step. Post-build, Peridot runs `createrepo_c` on all individual RPMs to create necessary metadata and stores that.

When an update request is sent to `yumrepofsupdater` containing a specific build, `GetBuildArtifacts` is called to fetch RPMs+Metadata.

Using this method, packages wouldn't need to coexist on disk to create a new repo, because that metadata could then be swapped into the correct place

Yumrepofsupdater first fetches the latest yumrepofs revision of that repo+arch, unmarshals it, then swaps in-place the metadata from the `GetBuildArtifacts` call.

Currently in-repo artifacts that are no longer to be found (based on NVRA), will be removed, just like with `createrepo_c`.

## Comps/groups

Using Peridot Catalogs, comps/groups can be declared that is later applied to every revision of target repo.

[comps2peridot](https://github.com/rocky-linux/peridot-releng/tree/main/comps2peridot) can be used to generate comps in a format Peridot accepts.

A configuration example can be found at [peridot-rocky](https://git.rockylinux.org/rocky/peridot-rocky/-/tree/r9)

## Sqlite/Hashed hrefs

Yumrepofs has something called "hashed variants" that can be created on demand. This changes package hrefs to the format of `Packages/{INITIAL}/{RPM_NAME}` (ex. `Packages/b/bash-5.1.8-4.el9.x86_64.rpm`).

This variant also runs `sqliterepo_c` on metadata and stores the sqlite files in S3. This is also served the same way as any yumrepofs artifact.
