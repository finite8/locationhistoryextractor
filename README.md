# locationhistoryextractor
Quick and nasty dump of location history data from google. Useful if you need to dig through your location history and try to figure out where you were at particular points in the day. As I primarily work from home, I put this together so I could figure out which days I was in the Office so I could figure out how much tax deductions i could claim in my tax return (gotta love tax time amirite?).

Excuse the hasty code and lack of tests. I had tests but broke them all with a refactor and CBF rewriting them too. It isn't perfect, but it gets the job done. 

To install (requires go 1.21 or newer. Would prob work with older too, but you should be up to date anyway. https://go.dev/doc/install):
```
go install github.com/finite8/locationhistoryextractor
```

## Getting location data from Google
1. Go to https://takeout.google.com
2. Press "Deselect all", you won't need all of it.
3. Scroll down until you find "Location History (Timeline)" and enable it, then scroll to the bottom and press "next"
4. Just do "Export Once". I also recommend in the "transfer to" you set it to your preferred cloud storage. Basically, wherever you can easily access the downloaded file from where you are going to run this tool.
5. Create the export and wait. Once done, just make sure you can access and open the zip file locally.

## Pulling out the data using this tool
The tool will work with either the original zip file OR, if you have already extracted it somewhere you can use that. It will recursively search the path/zip for files in the format of `[year]_*.json` and then it will deserialize it and pull out the "placesVisit" nodes and give you just the location Name, Address, TimeStart, TimeFinish and Duration.

Example:

```
locationhistoryextractor -source "C:\Users\[yourusername]\OneDrive\Apps\Google⁠ Download Your Data\takeout-20240504T231214Z-001.zip" -start "2022-07-01T00:00:00Z" -end "2023-06-30T00:00:000Z"
```
This will print out to the console the location data between the start and end date. Add ` > somefilename.csv` at the end of it if you want it dumped into a file.

## License

I wrote this up in a few hours and did a quick refactor in less than that. If you find any of this code useful, help yourself to it. 
