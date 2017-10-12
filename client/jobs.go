package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cozy/cozy-stack/client/request"
)

type jobOptions struct {
	MaxExecCount int            `json:"max_exec_count,omitempty"`
	MaxExecTime  *time.Duration `json:"max_exec_time,omitempty"`
	Timeout      *time.Duration `json:"timeout,omitempty"`
}

// JobOptions is the options to run a job.
type JobOptions struct {
	Worker       string
	Arguments    interface{}
	MaxExecCount int
	MaxExecTime  *time.Duration
	Timeout      *time.Duration
}

// Job is a struct representing a job
type Job struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Attrs struct {
		Domain    string          `json:"domain"`
		Message   json.RawMessage `json:"message"`
		Debounced bool            `json:"debounced"`
		Event     struct {
			Domain string          `json:"domain"`
			Verb   string          `json:"verb"`
			Doc    json.RawMessage `json:"doc"`
			OldDoc json.RawMessage `json:"old,omitempty"`
		} `json:"event"`
		Options   *jobOptions `json:"options"`
		QueuedAt  time.Time   `json:"queued_at"`
		StartedAt time.Time   `json:"started_at"`
		State     string      `json:"state"`
		Worker    string      `json:"worker"`
	} `json:"attributes"`
}

type Trigger struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Attrs struct {
		Domain     string          `json:"domain"`
		Type       string          `json:"type"`
		WorkerType string          `json:"worker"`
		Arguments  string          `json:"arguments"`
		Debounce   string          `json:"debounce"`
		Message    json.RawMessage `json:"message"`
		Options    *struct {
			MaxExecCount int           `json:"max_exec_count"`
			MaxExecTime  time.Duration `json:"max_exec_time"`
			Timeout      time.Duration `json:"timeout"`
		} `json:"options"`
	} `json:"attributes"`
}

// JobPush is used to push a new job into the job queue.
func (c *Client) JobPush(r *JobOptions) (*Job, error) {
	args, err := json.Marshal(r.Arguments)
	if err != nil {
		return nil, err
	}

	type jobAttrs struct {
		Arguments json.RawMessage `json:"arguments"`
		Options   *jobOptions     `json:"options"`
	}

	opt := &jobOptions{}
	if r.MaxExecCount > 0 {
		opt.MaxExecCount = r.MaxExecCount
	}
	if r.MaxExecTime != nil {
		opt.MaxExecTime = r.MaxExecTime
	}
	if r.Timeout != nil {
		opt.Timeout = r.Timeout
	}

	job := struct {
		Attrs jobAttrs `json:"attributes"`
	}{
		Attrs: jobAttrs{
			Arguments: args,
			Options:   opt,
		},
	}
	body, err := writeJSONAPI(job)
	if err != nil {
		return nil, err
	}
	res, err := c.Req(&request.Options{
		Method: "POST",
		Path:   "/jobs/queue/" + url.PathEscape(r.Worker),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	var j *Job
	if err := readJSONAPI(res.Body, &j); err != nil {
		return nil, err
	}
	return j, nil
}

func (c *Client) GetTrigger(triggerID string) (*Trigger, error) {
	res, err := c.Req(&request.Options{
		Method: "GET",
		Path:   fmt.Sprintf("/jobs/triggers/%s", url.PathEscape(triggerID)),
	})
	if err != nil {
		return nil, err
	}
	var t *Trigger
	if err := readJSONAPI(res.Body, &t); err != nil {
		return nil, err
	}
	return t, nil
}

func (c *Client) GetTriggers(worker string) ([]*Trigger, error) {
	res, err := c.Req(&request.Options{
		Method:  "GET",
		Path:    fmt.Sprintf("/jobs/triggers"),
		Queries: url.Values{"Worker": {worker}},
	})
	if err != nil {
		return nil, err
	}
	var t []*Trigger
	if err := readJSONAPI(res.Body, &t); err != nil {
		return nil, err
	}
	return t, nil
}

func (c *Client) TriggerRun(triggerID string) (*Job, error) {
	res, err := c.Req(&request.Options{
		Method: "POST",
		Path:   fmt.Sprintf("/jobs/triggers/%s/launch", url.PathEscape(triggerID)),
	})
	if err != nil {
		return nil, err
	}
	var j *Job
	if err := readJSONAPI(res.Body, &j); err != nil {
		return nil, err
	}
	return j, nil
}
