/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/beevik/etree"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	_ "github.com/beevik/etree"
	"github.com/vmware/go-vcloud-director/types/v56"
)

type VAppTemplate struct {
	VAppTemplate *types.VAppTemplate
	client       *Client
}

func NewVAppTemplate(cli *Client) *VAppTemplate {
	return &VAppTemplate{
		VAppTemplate: new(types.VAppTemplate),
		client:       cli,
	}
}

// GetTemplateDiskSize recovers disk size in bytes
func (vat *VAppTemplate) GetTemplateDiskSize() (int, error) {
	theUrl, _ := url.Parse(vat.VAppTemplate.HREF)

	req := vat.client.NewRequest(map[string]string{}, "GET", *theUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))

	for _, n := range xmlquery.Find(doc, "//ovf:Item") {
		desc := n.SelectElement("rasd:Description")
		if desc != nil {
			if strings.Contains(desc.InnerText(), "Hard disk") {
				diskSize := n.SelectElement("rasd:VirtualQuantity")
				if diskSize != nil {
					return strconv.Atoi(diskSize.InnerText())
				}
			}
		}
	}

	return 0, nil
}

// GetMemorySize recovers CPU size
func (vat *VAppTemplate) GetCPUSize() (int, error) {
	theUrl, _ := url.Parse(vat.VAppTemplate.HREF)

	req := vat.client.NewRequest(map[string]string{}, "GET", *theUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))

	for _, n := range xmlquery.Find(doc, "//ovf:Item") {
		desc := n.SelectElement("rasd:Description")
		if desc != nil {
			if strings.Contains(desc.InnerText(), "Number of Virtual CPUs") {
				diskSize := n.SelectElement("rasd:VirtualQuantity")
				if diskSize != nil {
					return strconv.Atoi(diskSize.InnerText())
				}
			}
		}
	}

	return 0, nil
}

// GetMemorySize recovers memory size in MB
func (vat *VAppTemplate) GetMemorySize() (int, error) {
	theUrl, _ := url.Parse(vat.VAppTemplate.HREF)

	req := vat.client.NewRequest(map[string]string{}, "GET", *theUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))

	for _, n := range xmlquery.Find(doc, "//ovf:Item") {
		desc := n.SelectElement("rasd:Description")
		if desc != nil {
			if strings.Contains(desc.InnerText(), "Memory Size") {
				diskSize := n.SelectElement("rasd:VirtualQuantity")
				if diskSize != nil {
					return strconv.Atoi(diskSize.InnerText())
				}
			}
		}
	}

	return 0, nil
}

func (vat *VAppTemplate) GetRaw(restUrl string) (string, error) {
	theUrl, err := url.Parse(vat.VAppTemplate.HREF + restUrl)
	if err != nil {
		return "", err
	}

	req := vat.client.NewRequest(map[string]string{}, "GET", *theUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (vat *VAppTemplate) PutRaw(restUrl string, content string) (Task, error) {
	theUrl, err := url.Parse(vat.VAppTemplate.HREF + restUrl)
	if err != nil {
		return Task{}, err
	}

	buffer := bytes.NewBufferString(content)

	req := vat.client.NewRequest(map[string]string{}, "PUT", *theUrl, buffer)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return Task{}, err
	}

	task := NewTask(vat.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

// getLinks gets vm links
func (vat *VAppTemplate) getLinks() (map[string]string, error) {
	theUrl, _ := url.Parse(vat.VAppTemplate.HREF)

	req := vat.client.NewRequest(map[string]string{}, "GET", *theUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vat.client.APIVersion)

	links := make(map[string]string)

	resp, err := checkResp(vat.client.Http.Do(req))
	if err != nil {
		return links, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return links, err
	}
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return links, err
	}

	dokk := etree.NewDocument()
	_ = dokk

	for _, n := range xmlquery.Find(doc, "//vcloud:Link") {
		href := n.SelectAttr("href")
		if href != "" {
			segments := strings.Split(href, "/")
			last := segments[len(segments)-1]
			links[last] = href
		}
	}

	return links, nil
}

func (vat *VAppTemplate) GetDiskLink() (string, error) {
	links, err := vat.getLinks()
	if err != nil {
		return "", err
	}

	return links["disks"], nil
}

func (vdc *Vdc) InstantiateVAppTemplate(template *types.InstantiateVAppTemplateParams) error {
	output, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}
	requestData := bytes.NewBufferString(xml.Header + string(output))

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	req := vdc.client.NewRequest(map[string]string{}, "POST", *vdcHref, requestData)
	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml")

	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error instantiating a new template: %s", err)
	}

	vapptemplate := NewVAppTemplate(vdc.client)
	if err = decodeBody(resp, vapptemplate.VAppTemplate); err != nil {
		return fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
	}
	task := NewTask(vdc.client)
	for _, taskItem := range vapptemplate.VAppTemplate.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error performing task: %#v", err)
		}
	}
	return nil
}
