# terraform-ops

## Usage

## `list-providers`

List the providers used in the given terraform workspaces.

```shell
$ terraform-ops list-providers <path1> <path2>

[
  {
    "path": "<path1>",
    "providers": {
      "google": "6.0.0",
      "google-beta": ">=6.0,<7.0"
    }
  },
  {
    "path": "<path2>",
    "providers": {
      "google": "6.0.0",
      "google-beta": ">=6.0,<7.0"
    }
  }
]
```
