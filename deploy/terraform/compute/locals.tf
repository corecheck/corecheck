locals {
  project_root_path = "${path.root}/../.."
  function_files = fileset("${local.project_root_path}/functions/compute", "**")
  function_file_hashes = concat(
    [
      filemd5("${local.project_root_path}/go.mod"),
      filemd5("${local.project_root_path}/go.sum"),
    ],
    [for f in local.function_files : filemd5("${local.project_root_path}/functions/compute/${f}") ],
  )
}
