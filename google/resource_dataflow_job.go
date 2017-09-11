package google

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"google.golang.org/api/dataflow/v1b3"
)

func resourceDataflowJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataflowJobCreate,
		Read:   resourceDataflowJobRead,
		Delete: resourceDataflowJobDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp(DataflowJobRegex),
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"template_gcs_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"temp_gcs_location": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"max_workers": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"on_delete": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"cancel", "drain"}, false),
				Optional:     true,
				Default:      "drain",
				ForceNew:     true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDataflowJobCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	gcsPath := d.Get("template_gcs_path").(string)
	tempLocation := d.Get("environment").(*schema.ResourceData).Get("template_gcs_location").(string)
	zone := d.Get("environment").(*schema.ResourceData).Get("zone").(string)
	maxWorkers := d.Get("environment").(*schema.ResourceData).Get("max_workers").(int)
	params := expandStringMap(d, "parameters")

	env := dataflow.RuntimeEnvironment{
		TempLocation: tempLocation,
		Zone:         zone,
		MaxWorkers:   int64(maxWorkers),
	}

	request := dataflow.CreateJobFromTemplateRequest{
		JobName:     name,
		GcsPath:     gcsPath,
		Parameters:  params,
		Environment: &env,
	}

	job, err := config.clientDataflow.Projects.Templates.Create(project, &request).Do()
	if err != nil {
		return err
	}

	d.SetId(job.Id)

	return resourceDataflowJobRead(d, meta)
}

func resourceDataflowJobRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	id := d.Id()

	job, err := config.clientDataflow.Projects.Jobs.Get(project, id).Do()
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Dataflow job %s", id))
	}

	d.Set("state", job.CurrentState)

	return nil
}

func resourceDataflowJobDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	id := d.Id()

	requestedState, err := mapOnDelete(d.Get("on_delete").(string))
	if err != nil {
		return err
	}

	job := &dataflow.Job{
		RequestedState: requestedState,
	}

	_, err = config.clientDataflow.Projects.Jobs.Update(project, id, job).Do()
	if err != nil {
		return err
	}

	return nil
}

func mapOnDelete(policy string) (string, error) {
	switch policy {
	case "cancel":
		return "JOB_STATE_CANCELLED", nil
	case "done":
		return "JOB_STATE_DONE", nil
	case "drain":
		return "JOB_STATE_DRAINING", nil
	default:
		return "", fmt.Errorf("Invalid `on_delete` policy: %s", policy)
	}
}
