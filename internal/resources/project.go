package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
	"github.com/kitokinha/terraform-provider-rustrak/internal/project"
)

func Project() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Rustrak project.",
			},
			"slug": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The slug of the Rustrak project.",
			},
			"platform": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The platform of the Rustrak project.",
			},
			"dsn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The DSN used to send events to the Rustrak project.",
			},
		},
	}
}

func resourceProjectCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	rustrakClient := meta.(*client.RustrakClient)

	proj := project.CreateProjectRequest{
		Name: d.Get("name").(string),
	}

	if slug, ok := d.GetOk("slug"); ok {
		proj.Slug = slug.(string)
	}

	if platform, ok := d.GetOk("platform"); ok {
		proj.Platform = platform.(string)
	}

	createdProject, err := project.CreateProject(
		rustrakClient,
		ctx,
		proj,
	)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("error creating project: %w", err),
		)
	}

	d.SetId(strconv.Itoa(createdProject.ID))

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	rustrakClient := meta.(*client.RustrakClient)

	project, err := project.ReadProject(
		rustrakClient,
		ctx,
		d.Id(),
	)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("error reading project %q: %w", d.Id(), err),
		)
	}

	if project == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("slug", project.Slug); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("platform", project.Platform); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("dsn", project.DSN); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProjectUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	rustrakClient := meta.(*client.RustrakClient)

	proj := project.UpdateProjectRequest{
		Name: d.Get("name").(string),
	}

	if slug, ok := d.GetOk("slug"); ok {
		proj.Slug = slug.(string)
	}

	if platform, ok := d.GetOk("platform"); ok {
		proj.Platform = platform.(string)
	}

	_, err := project.UpdateProject(
		rustrakClient,
		ctx,
		d.Id(),
		proj,
	)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("error updating project %q: %w", d.Id(), err),
		)
	}

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	rustrakClient := meta.(*client.RustrakClient)

	err := project.DeleteProject(
		rustrakClient,
		ctx,
		d.Id(),
	)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("error deleting project %q: %w", d.Id(), err),
		)
	}

	d.SetId("")

	return nil
}

func resourceProjectImportState(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil
}
