{{- $root := . -}}
{{- with .Data -}}

Title: {{ .Name }}
Template: gallery
Date: {{ .Date }}
Modified: {{ .Date }}
Status: hidden
Slug: {{ call $root.Slug .Name }}

<!-- TODO: buttons for filtering with isotope -->

<div class="grid" id="the-gallery" data-isotope='{"itemSelector": ".grid-item", "layoutMode": "masonry"}'>
{{ range .PiArr -}}
<div class="grid-item{{ if .IsPanorama }} grid-item-wide{{ end }}">
  <a href="{{ $root.DeployHref }}{{ .Filename }}"
    data-pswp-width="{{ .Width }}"
    data-pswp-height="{{ .Height }}"
    target="_blank">
  <img src="{{ $root.DeployHref }}{{ $root.ThumbnailsDir }}{{ .Filename }}" alt="" />
  <!-- TODO: generate description with tags and GPS coordinates -->
</a></div>
{{ end }}
</div>

{{- end -}}