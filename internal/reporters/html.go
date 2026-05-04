// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package reporters

import (
	"bytes"
	"fmt"
	"html"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

type HTMLReporter struct{}

type htmlReportSection struct {
	id         string
	label      string
	categories []string
}

func (r *HTMLReporter) Report(data map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Ampere System-Configurations Report</title>
<style>
*,
*::before,
*::after {
	box-sizing: border-box;
}

html {
	scroll-behavior: smooth;
}

:root {
	color-scheme: light;
	--bg: #f3f6fb;
	--panel: #ffffff;
	--panel-border: #dbe4ee;
	--panel-shadow: 0 20px 45px rgba(15, 23, 42, 0.08);
	--accent: #0f766e;
	--accent-soft: #ccfbf1;
	--ampere-red: #FF0000;
	--text: #0f172a;
	--muted: #64748b;
}

body {
	font-family: Arial, sans-serif;
	margin: 0;
	min-height: 100vh;
	line-height: 1.4;
	background:
		radial-gradient(circle at top left, rgba(15, 118, 110, 0.12), transparent 24rem),
		linear-gradient(180deg, #f8fbff 0%, var(--bg) 100%);
	color: var(--text);
	overflow: hidden;
}

h1 {
	margin: 0;
	font-size: 2rem;
	line-height: 1.1;
	color: var(--ampere-red);
}

h2 {
	margin-top: 0;
	margin-bottom: 1rem;
}

h3 {
	margin: 0 0 0.85rem;
	font-size: 1rem;
}

a {
	color: inherit;
}

.layout {
	display: grid;
	grid-template-columns: minmax(14rem, 17rem) minmax(0, 1fr);
	gap: 1.5rem;
	max-width: 96rem;
	margin: 0 auto;
	padding: 1.5rem;
	height: 100vh;
	align-items: stretch;
}

.sidebar {
	position: sticky;
	top: 0;
	align-self: start;
	background: rgba(255, 255, 255, 0.9);
	backdrop-filter: blur(14px);
	border: 1px solid var(--panel-border);
	border-radius: 16px;
	padding: 1rem;
	box-shadow: var(--panel-shadow);
	max-height: calc(100vh - 3rem);
	overflow: auto;
}

.sidebar-title {
	margin: 0 0 0.25rem;
	font-size: 1rem;
	font-weight: 700;
}

.sidebar-nav {
	display: grid;
	gap: 0.5rem;
}

.sidebar-link {
	display: block;
	padding: 0.65rem 0.8rem;
	border-radius: 10px;
	text-decoration: none;
	font-weight: 600;
	color: #134e4a;
	background: transparent;
	border: 1px solid transparent;
}

.sidebar-link:hover,
.sidebar-link:focus-visible {
	background: var(--accent-soft);
	border-color: rgba(15, 118, 110, 0.2);
	outline: none;
}

.content {
	min-width: 0;
	min-height: 0;
	display: grid;
	grid-template-rows: auto minmax(0, 1fr);
	gap: 1rem;
}

.hero {
	background: linear-gradient(135deg, rgba(15, 118, 110, 0.14), rgba(255, 255, 255, 0.94));
	border: 1px solid rgba(15, 118, 110, 0.18);
	border-radius: 20px;
	padding: 1.5rem;
	box-shadow: var(--panel-shadow);
}

.section-stack {
	min-height: 0;
	overflow-y: auto;
	padding-right: 0.35rem;
	scroll-behavior: smooth;
	scroll-padding-top: 1rem;
}

.section-group + .section-group {
	margin-top: 1.5rem;
	padding-top: 1.25rem;
	border-top: 1px solid var(--panel-border);
}

section {
	scroll-margin-top: 1rem;
	background: var(--panel);
	border: 1px solid var(--panel-border);
	border-radius: 16px;
	padding: 1.25rem;
	margin-bottom: 1rem;
	box-shadow: var(--panel-shadow);
}

section:target {
	border-color: rgba(15, 118, 110, 0.45);
	box-shadow:
		0 0 0 4px rgba(204, 251, 241, 0.9),
		var(--panel-shadow);
}

table {
	width: 100%;
	border-collapse: collapse;
	margin-top: 0.5rem;
}

th,
td {
	border: 1px solid var(--panel-border);
	padding: 0.5rem;
	text-align: left;
	vertical-align: top;
}

th {
	background: #eff6ff;
	font-weight: 600;
}

.definition-list {
	display: grid;
	grid-template-columns: minmax(12rem, 18rem) minmax(0, 1fr);
	margin-top: 0.5rem;
	border: 1px solid var(--panel-border);
	border-radius: 14px;
	overflow: hidden;
	background: #ffffff;
}

.definition-key,
.definition-value {
	margin: 0;
	padding: 0.75rem 0.9rem;
	border-top: 1px solid var(--panel-border);
}

.definition-list > :nth-child(-n+2) {
	border-top: none;
}

.definition-key {
	background: #f8fafc;
	font-weight: 600;
}

.definition-value {
	min-width: 0;
}

.definition-value > :first-child {
	margin-top: 0;
}

.collection-toggle {
	margin-top: 0.5rem;
	border: 1px solid var(--panel-border);
	border-radius: 12px;
	background: #f8fafc;
	overflow: hidden;
}

.collection-summary {
	cursor: pointer;
	padding: 0.75rem 0.9rem;
	font-weight: 600;
	color: #134e4a;
	background: rgba(204, 251, 241, 0.55);
}

.collection-summary::-webkit-details-marker {
	display: none;
}

.collection-summary::after {
	content: "+";
	float: right;
	color: var(--muted);
}

.collection-toggle[open] .collection-summary {
	border-bottom: 1px solid var(--panel-border);
}

.collection-toggle[open] .collection-summary::after {
	content: "-";
}

.collection-body {
	padding: 0 0.9rem 0.75rem;
}

.collection-body > ul,
.collection-body > ol {
	margin-top: 0.75rem;
}

.named-collection {
	display: grid;
	gap: 0.75rem;
	margin-top: 0.75rem;
}

.named-entry {
	border: 1px solid var(--panel-border);
	border-radius: 12px;
	background: #ffffff;
	overflow: hidden;
}

.named-entry-summary {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
	padding: 0.8rem 0.9rem;
	cursor: pointer;
	font-weight: 600;
}

.named-entry-summary::-webkit-details-marker {
	display: none;
}

.named-entry-summary::after {
	content: "+";
	color: var(--muted);
	flex: 0 0 auto;
}

.named-entry[open] .named-entry-summary::after {
	content: "-";
}

.named-entry-title {
	font-weight: 700;
	word-break: break-word;
}

.named-entry-meta {
	color: var(--muted);
	font-size: 0.9rem;
	font-weight: 500;
	text-align: right;
}

.named-entry-body {
	padding: 0 0.9rem 0.9rem;
	border-top: 1px solid var(--panel-border);
}

.named-entry-body > :first-child {
	margin-top: 0.75rem;
}

.dimm-summary {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
	gap: 0.75rem;
	margin-bottom: 1rem;
}

.dimm-stat {
	padding: 0.9rem 1rem;
	border: 1px solid rgba(15, 118, 110, 0.16);
	border-radius: 14px;
	background: linear-gradient(180deg, rgba(204, 251, 241, 0.7), rgba(255, 255, 255, 0.95));
}

.dimm-stat-label {
	display: block;
	margin-bottom: 0.35rem;
	color: var(--muted);
	font-size: 0.82rem;
	font-weight: 600;
	letter-spacing: 0.04em;
	text-transform: uppercase;
}

.dimm-stat-value {
	font-size: 1.5rem;
	font-weight: 700;
	line-height: 1;
}

.channel-panel {
	margin-bottom: 1rem;
	padding: 1rem;
	border: 1px solid var(--panel-border);
	border-radius: 14px;
	background: #f8fafc;
}

.channel-title {
	margin: 0 0 0.35rem;
	font-size: 0.95rem;
	font-weight: 700;
}

.channel-note {
	margin: 0 0 0.75rem;
	color: var(--muted);
	font-size: 0.9rem;
}

.channel-chip-list {
	display: flex;
	flex-wrap: wrap;
	gap: 0.55rem;
}

.channel-matrix {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
	gap: 0.85rem;
}

.channel-card {
	border: 1px solid rgba(15, 118, 110, 0.16);
	border-radius: 14px;
	background: #ffffff;
	padding: 0.9rem;
	box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.channel-card-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
	margin-bottom: 0.75rem;
}

.channel-card-count {
	color: var(--muted);
	font-size: 0.85rem;
	font-weight: 600;
}

.channel-chip {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	min-width: 2.6rem;
	padding: 0.45rem 0.8rem;
	border-radius: 999px;
	background: var(--accent);
	color: #ffffff;
	font-weight: 700;
	letter-spacing: 0.04em;
	box-shadow: 0 8px 20px rgba(15, 118, 110, 0.18);
}

.channel-slot-list {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
}

.channel-slot {
	display: inline-flex;
	align-items: center;
	padding: 0.42rem 0.65rem;
	border-radius: 10px;
	background: #eef2ff;
	color: #1e3a8a;
	font-size: 0.9rem;
	font-weight: 600;
}

ul,
ol {
	margin: 0.5rem 0 0.5rem 1.25rem;
	padding: 0;
}

.scalar {
	white-space: pre-wrap;
	word-break: break-word;
}

.empty {
	color: var(--muted);
	font-style: italic;
}

@media (max-width: 900px) {
	body {
		overflow: auto;
	}

	.layout {
		grid-template-columns: 1fr;
		height: auto;
		min-height: 100vh;
	}

	.sidebar {
		position: static;
		max-height: none;
		overflow: visible;
	}

	.sidebar-nav {
		grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
	}

	.content {
		display: block;
	}

	.hero {
		margin-bottom: 1rem;
	}

	.section-stack {
		overflow: visible;
		padding-right: 0;
	}
}

@media (max-width: 640px) {
	.definition-list {
		grid-template-columns: 1fr;
	}
}
</style>
</head>
<body>
<div class="layout">
<aside class="sidebar" aria-label="Subsystem navigation">
<p class="sidebar-title">Subsystems</p>
<nav class="sidebar-nav">
`)

	sections := htmlReportSections(data)

	for _, section := range sections {
		buf.WriteString(`<a class="sidebar-link" href="#`)
		buf.WriteString(section.id)
		buf.WriteString(`">`)
		buf.WriteString(html.EscapeString(section.label))
		buf.WriteString("</a>\n")
	}

	buf.WriteString(`</nav>
</aside>
<main class="content">
<header class="hero">
<h1>Ampere Config Reporter</h1>
</header>
<div class="section-stack">
`)

	for _, section := range sections {
		buf.WriteString(`<section id="`)
		buf.WriteString(section.id)
		buf.WriteString(`">` + "\n<h2>")
		buf.WriteString(html.EscapeString(section.label))
		buf.WriteString("</h2>\n")
		renderHTMLSectionCategories(&buf, data, section.categories)
		buf.WriteString("\n</section>\n")
	}

	buf.WriteString("</div>\n</main>\n</div>\n</body>\n</html>\n")
	return buf.Bytes(), nil
}

type mapEntry struct {
	key   string
	value reflect.Value
}

type dimmChannelGroup struct {
	channel string
	slots   []string
}

const collapsibleNestedListThreshold = 3

func renderHTMLCategoryValue(buf *bytes.Buffer, category string, value reflect.Value) {
	if category == "dimm_info" {
		renderHTMLDIMMSection(buf, value)
		return
	}
	renderHTMLValue(buf, value)
}

func renderHTMLSectionCategories(buf *bytes.Buffer, data map[string]interface{}, categories []string) {
	available := make([]string, 0, len(categories))
	for _, category := range categories {
		if _, ok := data[category]; ok {
			available = append(available, category)
		}
	}

	if len(available) == 0 {
		buf.WriteString(`<span class="empty">null</span>`)
		return
	}

	for _, category := range available {
		if len(available) > 1 {
			buf.WriteString(`<div class="section-group"><h3>`)
			buf.WriteString(html.EscapeString(htmlGroupedCategoryLabel(category)))
			buf.WriteString(`</h3>`)
		}

		renderHTMLCategoryValue(buf, category, reflect.ValueOf(data[category]))

		if len(available) > 1 {
			buf.WriteString(`</div>`)
		}
	}
}

func renderHTMLValue(buf *bytes.Buffer, value reflect.Value) {
	value = unwrapHTMLValue(value)
	if !value.IsValid() {
		buf.WriteString(`<span class="empty">null</span>`)
		return
	}

	switch value.Kind() {
	case reflect.Map:
		renderHTMLMap(buf, value)
	case reflect.Slice, reflect.Array:
		renderHTMLSlice(buf, value)
	default:
		buf.WriteString(`<span class="scalar">`)
		buf.WriteString(html.EscapeString(fmt.Sprint(value.Interface())))
		buf.WriteString(`</span>`)
	}
}

func renderHTMLDIMMSection(buf *bytes.Buffer, value reflect.Value) {
	value = unwrapHTMLValue(value)
	if !value.IsValid() || value.Kind() != reflect.Map {
		renderHTMLValue(buf, value)
		return
	}

	renderDIMMChannelSummary(buf, value)
	renderHTMLMapFiltered(buf, value, map[string]bool{
		"detected_channel_names": true,
	})
}

func renderHTMLMap(buf *bytes.Buffer, value reflect.Value) {
	renderHTMLMapFiltered(buf, value, nil)
}

func renderHTMLMapFiltered(buf *bytes.Buffer, value reflect.Value, hiddenKeys map[string]bool) {
	if value.Len() == 0 {
		buf.WriteString(`<span class="empty">{}</span>`)
		return
	}

	entries := sortedMapEntriesFiltered(value, hiddenKeys)
	if len(entries) == 0 {
		buf.WriteString(`<span class="empty">{}</span>`)
		return
	}

	buf.WriteString(`<dl class="definition-list">`)
	for _, entry := range entries {
		buf.WriteString(`<dt class="definition-key">`)
		buf.WriteString(html.EscapeString(entry.key))
		buf.WriteString(`</dt><dd class="definition-value">`)
		renderHTMLMapEntryValue(buf, entry.key, entry.value)
		buf.WriteString(`</dd>`)
	}
	buf.WriteString(`</dl>`)
}

func renderHTMLMapEntryValue(buf *bytes.Buffer, key string, value reflect.Value) {
	if renderHTMLNamedMapCollection(buf, key, value) {
		return
	}
	if renderHTMLNamedSliceCollection(buf, key, value) {
		return
	}
	renderHTMLValue(buf, value)
}

func renderHTMLSlice(buf *bytes.Buffer, value reflect.Value) {
	if value.Len() == 0 {
		buf.WriteString(`<span class="empty">[]</span>`)
		return
	}

	listTag := "ol"
	isScalar := isScalarSequence(value)
	if isScalar {
		listTag = "ul"
	}

	if !isScalar && value.Len() > collapsibleNestedListThreshold {
		renderHTMLCollapsibleSlice(buf, value, listTag)
		return
	}

	renderHTMLSequenceList(buf, value, listTag)
}

func renderHTMLCollapsibleSlice(buf *bytes.Buffer, value reflect.Value, listTag string) {
	itemLabel := "items"
	if value.Len() == 1 {
		itemLabel = "item"
	}

	buf.WriteString(`<details class="collection-toggle"><summary class="collection-summary">Show `)
	buf.WriteString(html.EscapeString(fmt.Sprintf("%d %s", value.Len(), itemLabel)))
	buf.WriteString(`</summary><div class="collection-body">`)
	renderHTMLSequenceList(buf, value, listTag)
	buf.WriteString(`</div></details>`)
}

func renderHTMLSequenceList(buf *bytes.Buffer, value reflect.Value, listTag string) {
	buf.WriteString("<")
	buf.WriteString(listTag)
	buf.WriteString(">\n")
	for i := 0; i < value.Len(); i++ {
		buf.WriteString("<li>")
		renderHTMLValue(buf, value.Index(i))
		buf.WriteString("</li>\n")
	}
	buf.WriteString("</")
	buf.WriteString(listTag)
	buf.WriteString(">")
}

func renderHTMLNamedSliceCollection(buf *bytes.Buffer, key string, value reflect.Value) bool {
	var singularLabel string
	var pluralLabel string
	var titleFunc func(reflect.Value, int) (string, map[string]bool)
	var summaryFunc func(reflect.Value) string

	switch key {
	case "containers":
		singularLabel = "container"
		pluralLabel = "containers"
		titleFunc = htmlContainerEntryTitle
		summaryFunc = htmlContainerEntrySummary
	case "devices":
		singularLabel = "PCIe device"
		pluralLabel = "PCIe devices"
		titleFunc = htmlPCIeDeviceEntryTitle
		summaryFunc = htmlPCIeDeviceEntrySummary
	default:
		return false
	}

	value = unwrapHTMLValue(value)
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) || value.Len() == 0 {
		return false
	}

	items := make([]reflect.Value, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		item := unwrapHTMLValue(value.Index(i))
		if !item.IsValid() || item.Kind() != reflect.Map {
			return false
		}
		items = append(items, item)
	}

	itemLabel := pluralLabel
	if len(items) == 1 {
		itemLabel = singularLabel
	}

	buf.WriteString(`<details class="collection-toggle"><summary class="collection-summary">Show `)
	buf.WriteString(html.EscapeString(fmt.Sprintf("%d %s", len(items), itemLabel)))
	buf.WriteString(`</summary><div class="collection-body"><div class="named-collection">`)
	for i, item := range items {
		title, hiddenKeys := titleFunc(item, i)
		buf.WriteString(`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">`)
		buf.WriteString(html.EscapeString(title))
		buf.WriteString(`</span>`)
		if summary := summaryFunc(item); summary != "" {
			buf.WriteString(`<span class="named-entry-meta">`)
			buf.WriteString(html.EscapeString(summary))
			buf.WriteString(`</span>`)
		}
		buf.WriteString(`</summary><div class="named-entry-body">`)
		renderHTMLMapFiltered(buf, item, hiddenKeys)
		buf.WriteString(`</div></details>`)
	}
	buf.WriteString(`</div></div></details>`)
	return true
}

func renderHTMLNamedMapCollection(buf *bytes.Buffer, key string, value reflect.Value) bool {
	var singularLabel string
	var pluralLabel string
	var summaryFunc func(reflect.Value) string

	switch key {
	case "interfaces":
		singularLabel = "interface"
		pluralLabel = "interfaces"
		summaryFunc = htmlInterfaceEntrySummary
	case "partitions":
		singularLabel = "partition"
		pluralLabel = "partitions"
		summaryFunc = htmlPartitionEntrySummary
	default:
		return false
	}

	value = unwrapHTMLValue(value)
	if !value.IsValid() || value.Kind() != reflect.Map || value.Len() == 0 {
		return false
	}

	entries := sortedMapEntries(value)
	for _, entry := range entries {
		item := unwrapHTMLValue(entry.value)
		if !item.IsValid() || item.Kind() != reflect.Map {
			return false
		}
	}

	itemLabel := pluralLabel
	if len(entries) == 1 {
		itemLabel = singularLabel
	}

	buf.WriteString(`<details class="collection-toggle"><summary class="collection-summary">Show `)
	buf.WriteString(html.EscapeString(fmt.Sprintf("%d %s", len(entries), itemLabel)))
	buf.WriteString(`</summary><div class="collection-body"><div class="named-collection">`)
	for _, entry := range entries {
		buf.WriteString(`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">`)
		buf.WriteString(html.EscapeString(entry.key))
		buf.WriteString(`</span>`)
		if summary := summaryFunc(entry.value); summary != "" {
			buf.WriteString(`<span class="named-entry-meta">`)
			buf.WriteString(html.EscapeString(summary))
			buf.WriteString(`</span>`)
		}
		buf.WriteString(`</summary><div class="named-entry-body">`)
		renderHTMLValue(buf, entry.value)
		buf.WriteString(`</div></details>`)
	}
	buf.WriteString(`</div></div></details>`)
	return true
}

func isScalarSequence(value reflect.Value) bool {
	for i := 0; i < value.Len(); i++ {
		item := unwrapHTMLValue(value.Index(i))
		if !item.IsValid() {
			continue
		}

		switch item.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array:
			return false
		}
	}
	return true
}

func renderDIMMChannelSummary(buf *bytes.Buffer, value reflect.Value) {
	channels := htmlStringSequence(mapValueByStringKey(value, "detected_channel_names"))
	if len(channels) == 0 {
		return
	}

	slotCount := fmt.Sprint(valueFromMapOrFallback(value, "total_memory_slots_populated", len(channels)))
	channelCount := fmt.Sprint(valueFromMapOrFallback(value, "detected_memory_channels", len(channels)))

	buf.WriteString(`<div class="dimm-summary">`)
	renderDIMMStat(buf, "Populated Slots", slotCount)
	renderDIMMStat(buf, "Populated Channels", channelCount)
	buf.WriteString(`</div>`)

	buf.WriteString(`<div class="channel-panel" aria-label="Populated memory channels">`)
	buf.WriteString(`<p class="channel-title">Populated Channels</p>`)
	channelGroups := buildDIMMChannelGroups(mapValueByStringKey(value, "populated_dimm_details"))
	if len(channelGroups) > 0 {
		buf.WriteString(`<p class="channel-note">Grouped by inferred channel, with populated slot locators shown for each channel.</p>`)
		buf.WriteString(`<div class="channel-matrix">`)
		for _, group := range channelGroups {
			buf.WriteString(`<div class="channel-card"><div class="channel-card-head">`)
			buf.WriteString(`<span class="channel-chip">`)
			buf.WriteString(html.EscapeString(group.channel))
			buf.WriteString(`</span><span class="channel-card-count">`)
			buf.WriteString(html.EscapeString(fmt.Sprintf("%d DIMMs", len(group.slots))))
			buf.WriteString(`</span></div><div class="channel-slot-list">`)
			for _, slot := range group.slots {
				buf.WriteString(`<span class="channel-slot">`)
				buf.WriteString(html.EscapeString(slot))
				buf.WriteString(`</span>`)
			}
			buf.WriteString(`</div></div>`)
		}
		buf.WriteString(`</div>`)
	} else {
		buf.WriteString(`<p class="channel-note">Channels inferred from DIMM locator data.</p>`)
		buf.WriteString(`<div class="channel-chip-list">`)
		for _, channel := range channels {
			buf.WriteString(`<span class="channel-chip">`)
			buf.WriteString(html.EscapeString(channel))
			buf.WriteString(`</span>`)
		}
		buf.WriteString(`</div>`)
	}
	buf.WriteString(`</div>`)
}

func renderDIMMStat(buf *bytes.Buffer, label string, value string) {
	buf.WriteString(`<div class="dimm-stat"><span class="dimm-stat-label">`)
	buf.WriteString(html.EscapeString(label))
	buf.WriteString(`</span><span class="dimm-stat-value">`)
	buf.WriteString(html.EscapeString(value))
	buf.WriteString(`</span></div>`)
}

func mapValueByStringKey(value reflect.Value, key string) reflect.Value {
	value = unwrapHTMLValue(value)
	if !value.IsValid() || value.Kind() != reflect.Map {
		return reflect.Value{}
	}
	return unwrapHTMLValue(value.MapIndex(reflect.ValueOf(key)))
}

func valueFromMapOrFallback(value reflect.Value, key string, fallback int) interface{} {
	v := mapValueByStringKey(value, key)
	if !v.IsValid() {
		return fallback
	}
	return v.Interface()
}

func htmlStringSequence(value reflect.Value) []string {
	value = unwrapHTMLValue(value)
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) {
		return nil
	}

	items := make([]string, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		item := unwrapHTMLValue(value.Index(i))
		if !item.IsValid() {
			continue
		}
		items = append(items, fmt.Sprint(item.Interface()))
	}
	return items
}

func htmlContainerEntryTitle(value reflect.Value, index int) (string, map[string]bool) {
	title := stringFromMapKey(value, "Name")
	if title == "" {
		title = fmt.Sprintf("Container %d", index+1)
		return title, nil
	}
	return title, map[string]bool{
		"Name": true,
	}
}

func htmlContainerEntrySummary(value reflect.Value) string {
	parts := make([]string, 0, 2)

	if image := stringFromMapKey(value, "Image"); image != "" {
		parts = append(parts, image)
	}

	if status := stringFromMapKey(value, "Status"); status != "" {
		parts = append(parts, status)
	}

	return strings.Join(parts, " · ")
}

func htmlInterfaceEntrySummary(value reflect.Value) string {
	parts := make([]string, 0, 3)

	if state := stringFromMapKey(value, "state"); state != "" {
		parts = append(parts, state)
	}

	if speed := stringFromMapKey(value, "speed"); speed != "" {
		parts = append(parts, speed)
	}

	ipv4Count := len(htmlStringSequence(mapValueByStringKey(value, "ipv4_addresses")))
	if ipv4Count > 0 {
		label := "IPv4 addresses"
		if ipv4Count == 1 {
			label = "IPv4 address"
		}
		parts = append(parts, fmt.Sprintf("%d %s", ipv4Count, label))
	}

	return strings.Join(parts, " · ")
}

func htmlPCIeDeviceEntryTitle(value reflect.Value, index int) (string, map[string]bool) {
	title := stringFromMapKey(value, "Device")
	if title != "" {
		return title, map[string]bool{
			"Device": true,
		}
	}

	title = stringFromMapKey(value, "Slot")
	if title != "" {
		return title, map[string]bool{
			"Slot": true,
		}
	}

	return fmt.Sprintf("PCIe Device %d", index+1), nil
}

func htmlPCIeDeviceEntrySummary(value reflect.Value) string {
	parts := make([]string, 0, 3)

	if slot := stringFromMapKey(value, "Slot"); slot != "" {
		parts = append(parts, slot)
	}

	if vendor := stringFromMapKey(value, "Vendor"); vendor != "" {
		parts = append(parts, vendor)
	}

	if class := stringFromMapKey(value, "Class"); class != "" {
		parts = append(parts, class)
	}

	return strings.Join(parts, " · ")
}

func htmlPartitionEntrySummary(value reflect.Value) string {
	parts := make([]string, 0, 2)

	if device := stringFromMapKey(value, "device"); device != "" {
		parts = append(parts, device)
	}

	if fstype := stringFromMapKey(value, "fstype"); fstype != "" {
		parts = append(parts, fstype)
	}

	return strings.Join(parts, " · ")
}

func buildDIMMChannelGroups(value reflect.Value) []dimmChannelGroup {
	value = unwrapHTMLValue(value)
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) {
		return nil
	}

	grouped := make(map[string][]string)
	for i := 0; i < value.Len(); i++ {
		item := unwrapHTMLValue(value.Index(i))
		if !item.IsValid() || item.Kind() != reflect.Map {
			continue
		}

		locator := stringFromMapKey(item, "Locator")
		bankLocator := stringFromMapKey(item, "Bank Locator")
		size := stringFromMapKey(item, "Size")
		channel := inferDIMMChannel(locator, bankLocator)
		if channel == "" {
			continue
		}

		slotLabel := strings.TrimSpace(locator)
		if slotLabel == "" {
			slotLabel = strings.TrimSpace(bankLocator)
		}
		if slotLabel == "" {
			slotLabel = fmt.Sprintf("DIMM %d", i+1)
		}
		if size != "" {
			slotLabel = fmt.Sprintf("%s · %s", slotLabel, size)
		}

		grouped[channel] = append(grouped[channel], slotLabel)
	}

	channels := make([]string, 0, len(grouped))
	for channel := range grouped {
		channels = append(channels, channel)
	}
	sort.Strings(channels)

	groups := make([]dimmChannelGroup, 0, len(channels))
	for _, channel := range channels {
		slots := grouped[channel]
		sort.Strings(slots)
		groups = append(groups, dimmChannelGroup{
			channel: channel,
			slots:   slots,
		})
	}

	return groups
}

func stringFromMapKey(value reflect.Value, key string) string {
	item := mapValueByStringKey(value, key)
	if !item.IsValid() {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(item.Interface()))
}

func inferDIMMChannel(locator string, bankLocator string) string {
	for _, candidate := range []string{bankLocator, locator} {
		channel := strings.ToUpper(strings.TrimSpace(candidate))
		if channel == "" {
			continue
		}

		if idx := strings.Index(channel, "CH "); idx >= 0 {
			if token := leadingDIMMToken(channel[idx+3:]); token != "" {
				return token
			}
		}
		if idx := strings.Index(channel, "CHANNEL"); idx >= 0 {
			if token := leadingDIMMToken(channel[idx+len("CHANNEL"):]); token != "" {
				return token
			}
		}
		if idx := strings.Index(channel, "DIMM_P"); idx >= 0 {
			parts := strings.Split(channel[idx:], "_")
			if len(parts) >= 2 && parts[1] != "" {
				runes := []rune(parts[1])
				if len(runes) > 0 {
					return string(runes[0])
				}
			}
		}
		if idx := strings.Index(channel, "DIMM"); idx >= 0 {
			remainder := strings.TrimPrefix(channel[idx:], "DIMM")
			remainder = strings.TrimSpace(remainder)
			if remainder != "" {
				runes := []rune(remainder)
				if len(runes) > 0 && unicode.IsLetter(runes[0]) {
					return string(runes[0])
				}
			}
		}
	}

	return ""
}

func leadingDIMMToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var token []rune
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			token = append(token, r)
			continue
		}
		if len(token) > 0 {
			break
		}
	}
	return string(token)
}

func sortedMapEntries(value reflect.Value) []mapEntry {
	return sortedMapEntriesFiltered(value, nil)
}

func sortedMapEntriesFiltered(value reflect.Value, hiddenKeys map[string]bool) []mapEntry {
	entries := make([]mapEntry, 0, value.Len())
	iter := value.MapRange()
	for iter.Next() {
		key := fmt.Sprint(iter.Key().Interface())
		if hiddenKeys != nil && hiddenKeys[key] {
			continue
		}
		entries = append(entries, mapEntry{
			key:   key,
			value: iter.Value(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	return entries
}

func unwrapHTMLValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func htmlSectionID(category string) string {
	var builder strings.Builder
	builder.WriteString("section-")

	lastHyphen := true
	for _, r := range strings.ToLower(category) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastHyphen = false
			continue
		}

		if !lastHyphen {
			builder.WriteByte('-')
			lastHyphen = true
		}
	}

	sectionID := strings.TrimSuffix(builder.String(), "-")
	if sectionID == "section" {
		return "section-report"
	}
	return sectionID
}

func htmlCategoryLabel(category string) string {
	base := strings.TrimSuffix(category, "_info")
	switch base {
	case "cpu":
		return "CPU"
	case "mem":
		return "Memory"
	}

	parts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '_' || r == '-'
	})
	if len(parts) == 0 {
		return category
	}

	for i, part := range parts {
		parts[i] = htmlCategoryWord(part)
	}

	return strings.Join(parts, " ")
}

func htmlCategoryWord(word string) string {
	switch strings.ToLower(word) {
	case "cpu":
		return "CPU"
	case "os":
		return "OS"
	case "pcie":
		return "PCIe"
	case "pmu":
		return "PMU"
	}

	if word == "" {
		return word
	}

	runes := []rune(strings.ToLower(word))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func htmlCategoryOrderIndex(category string) int {
	switch category {
	case "system_info":
		return 0
	case "cpu_info":
		return 1
	case "dimm_info":
		return 2
	case "mem_info":
		return 3
	case "disk_info":
		return 4
	case "network_info":
		return 5
	case "software_info":
		return 6
	case "pmu_info":
		return 7
	default:
		return 100
	}
}

func htmlReportSections(data map[string]interface{}) []htmlReportSection {
	sections := make([]htmlReportSection, 0, 7)
	known := []htmlReportSection{
		{id: htmlSectionID("system_info"), label: "System", categories: []string{"system_info"}},
		{id: htmlSectionID("cpu_info"), label: "CPU", categories: []string{"cpu_info"}},
		{id: htmlSectionID("memory_info"), label: "Memory", categories: []string{"mem_info", "dimm_info"}},
		{id: htmlSectionID("disk_info"), label: "Disk", categories: []string{"disk_info"}},
		{id: htmlSectionID("network_info"), label: "Network", categories: []string{"network_info"}},
		{id: htmlSectionID("software_info"), label: "Software", categories: []string{"software_info"}},
		{id: htmlSectionID("pmu_info"), label: "PMU", categories: []string{"pmu_info"}},
	}

	for _, section := range known {
		for _, category := range section.categories {
			if _, ok := data[category]; ok {
				sections = append(sections, section)
				break
			}
		}
	}

	extraCategories := make([]string, 0)
	for category := range data {
		if category == "mem_info" || category == "dimm_info" {
			continue
		}
		if htmlCategoryOrderIndex(category) != 100 {
			continue
		}
		extraCategories = append(extraCategories, category)
	}
	sort.Strings(extraCategories)

	for _, category := range extraCategories {
		sections = append(sections, htmlReportSection{
			id:         htmlSectionID(category),
			label:      htmlCategoryLabel(category),
			categories: []string{category},
		})
	}

	return sections
}

func htmlGroupedCategoryLabel(category string) string {
	switch category {
	case "mem_info":
		return "Memory Info"
	case "dimm_info":
		return "DIMM Info"
	default:
		return htmlCategoryLabel(category)
	}
}
