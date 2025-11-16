package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Url       string `json:"url"`
	Clickable bool   `json:"clickable"`
}

var tmpl *template.Template

func main() {
	// load template
	var err error
	tmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}

	// port from env, default 8080
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	http.HandleFunc("/", handleHtml)
	http.HandleFunc("/api/containers", handleJson)

	addr := ":" + strconv.Itoa(port)
	log.Printf("Parlabuhan running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getContainers() ([]ContainerItem, error) {
	ctx := context.Background()

	// Create client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// Get container list
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var list []ContainerItem

	for _, c := range containers {

		if c.Labels["container.hidden"] == "true" {
			continue
		}

		// Ambil nama
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		item := ContainerItem{
			ID:    c.ID[:12],
			Name:  name,
			Image: c.Image,
		}

		if c.Labels["link.expose"] == "true" {
			item.Url = "http://" + name + ".localhost"
			item.Clickable = true
		}

		list = append(list, item)
	}

	return list, nil
}

func handleHtml(w http.ResponseWriter, r *http.Request) {
	containers, err := getContainers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.Execute(w, containers); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func handleJson(w http.ResponseWriter, r *http.Request) {
	containers, err := getContainers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}
