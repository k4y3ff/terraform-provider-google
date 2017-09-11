---
layout: "google"
page_title: "Google: google_dataflow_job"
sidebar_current: "docs-google-dataflow-job"
description: |-
  Creates a Google Cloud Dataflow job.
---

# google\_dataflow\_job

Creates a Google Cloud Dataflow job. For more information see [the official documentation](https://cloud.google.com/dataflow/docs/) and [API](https://cloud.google.com/dataflow/docs/reference/rest/).

## Example Usage

```hcl
resource "google_dataflow_job" "default" {
  name = "default-dataflow-job"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the resource, required by Dataflow. Changing this forces a new resource to be created.
