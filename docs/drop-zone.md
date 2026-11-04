# Drop zone

A well with a dashed border that takes files from other apps, Finder or
Explorer: when files come over it, the border turns solid and takes the
accent. It returns the paths dropped, in that frame. `FileItem` is the row
of a file added, with a button taking it out.

<!-- preview drop-zone 270 -->

```go
for _, path := range mizan.DropZone(c.Key("upload"), mizan.DropZoneOptions{Title: "Drop files here", Hint: "PDF, XLSX · up to 10 MB"}) {
	app.add(path)
}
for i, f := range app.files {
	if mizan.FileItem(c.Key(f.Name), f.Name, f.Size) {
		app.remove(i)
	}
}
```

With `Browse`, the zone takes a click too, and calls it: open a file
dialog there, as `mygo.Dialog.Open`.
