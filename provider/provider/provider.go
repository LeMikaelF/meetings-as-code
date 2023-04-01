package provider

import (
	"context"
	"github.com/LeMikaelF/meetings-as-code/provider/graphapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GRAPH_API_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"graphapi_calendar_event": resourceCalendarEvent(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	client := graphapi.NewGraphAPIClient(token)

	return client, nil
}

func resourceCalendarEvent() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEvent,
		ReadContext:   readEvent,
		UpdateContext: updateEvent,
		DeleteContext: deleteEvent,
		Schema: map[string]*schema.Schema{
			"subject": {
				Type:     schema.TypeString,
				Required: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attendee": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"required", "optional"}, false),
						},
					},
				},
			},
			"show_as": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"free", "tentative", "busy", "oof"}, false),
			},
		},
	}
}

func createEvent(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphapi.GraphAPIClient)
	event, diags := resourceDataToEvent(d)
	if diags.HasError() {
		return diags
	}

	createdEvent, err := client.CreateEvent(ctx, *event)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdEvent.ID)
	return nil
}

func readEvent(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphapi.GraphAPIClient)
	eventID := d.Id()

	evs, err := client.ReadEvents(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var event *graphapi.Event
	for _, e := range evs {
		if e.ID == eventID {
			event = &e
			break
		}
	}

	if event == nil {
		d.SetId("")
		return nil
	}

	d.Set("subject", event.Subject)
	d.Set("start_time", event.StartTime.DateTime)
	d.Set("end_time", event.EndTime.DateTime)
	d.Set("time_zone", event.StartTime.TimeZone)
	d.Set("location", event.Location.DisplayName)
	d.Set("attendee", flattenAttendees(event.Attendees))
	d.Set("show_as", event.ShowAs)

	return nil
}

func updateEvent(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphapi.GraphAPIClient)
	eventID := d.Id()

	event, diags := resourceDataToEvent(d)
	if diags.HasError() {
		return diags
	}

	err := client.UpdateEvent(ctx, eventID, *event)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func deleteEvent(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphapi.GraphAPIClient)
	eventID := d.Id()

	err := client.DeleteEvent(ctx, eventID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDataToEvent(d *schema.ResourceData) (*graphapi.Event, diag.Diagnostics) {
	var diags diag.Diagnostics
	event := &graphapi.Event{
		Subject: d.Get("subject").(string),
		StartTime: graphapi.DateTime{
			DateTime: d.Get("start_time").(string),
			TimeZone: d.Get("time_zone").(string),
		},
		EndTime: graphapi.DateTime{
			DateTime: d.Get("end_time").(string),
			TimeZone: d.Get("time_zone").(string),
		},
		Location: graphapi.Location{
			DisplayName: d.Get("location").(string),
		},
		Attendees: expandAttendees(d.Get("attendee").(*schema.Set).List()),
		ShowAs:    d.Get("show_as").(string),
	}

	return event, diags
}

func expandAttendees(attendees []interface{}) []graphapi.Attendee {
	result := make([]graphapi.Attendee, len(attendees))

	for i, a := range attendees {
		m := a.(map[string]interface{})
		result[i] = graphapi.Attendee{
			Type: m["type"].(string),
			EmailAddress: graphapi.EmailAddress{
				Name:    m["name"].(string),
				Address: m["address"].(string),
			},
		}
	}

	return result
}

func flattenAttendees(attendees []graphapi.Attendee) []map[string]interface{} {
	result := make([]map[string]interface{}, len(attendees))

	for i, a := range attendees {
		result[i] = map[string]interface{}{
			"type":    a.Type,
			"name":    a.EmailAddress.Name,
			"address": a.EmailAddress.Address,
		}
	}

	return result
}
