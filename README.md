# gribdl - Downloading Weather Data made easy

gribdl makes it easy and to download weather data fast from the [Deutscher Wetterdienst](https://www.dwd.de/EN/Home/home_node.html) (DWD) and the [National Oceanic and Atmospheric Administration](https://www.noaa.gov/) (NOAA).

# Supported Weather Models
- DWD ICON-EU
- DWD ICON-D2
- DWD ICON
- NOAA GFS
- NOAA NAM (not yet implemented)

You can read more about the data sources [here](https://www.dwd.de/EN/ourservices/opendata/opendata.html) and [here](https://www.ncdc.noaa.gov/data-access/model-data/model-datasets/global-forcast-system-gfs).

# Features
- HTTP/2 support for faster downloads
- Parallel downloads
- Setting max forecast hours
- Downloading multiple parameters at once


# Usage

Printing help:
```bash
docker run --rm -v ${PWD}/output:/app/output ghcr.io/hstin-de/gribdl --help
```

Downloading 8h T_2M from ICON-EU:
```bash
docker run --rm -v ${PWD}/output:/app/output ghcr.io/hstin-de/gribdl dwd icon-eu --param=T_2M --maxStep=8
```

Downloading 8h TMP from GFP:
```bash
docker run --rm -v ${PWD}/output:/app/output ghcr.io/hstin-de/gribdl noaa gfs --param=TMP --maxStep=8
```

# Building


```bash
git clone https://github.com/hstin-de/gribdl.git
cd gribdl
```

```bash
docker build -t hstin-de/gribdl .
```





# Usage Without Docker

You can also run gribdl without docker but its not recommended as
all the dependencies are included within the docker image.

### Prequisites:
- go 1.21.5
- [cdo 1.9.10 with netcdf and grib2 support](https://gist.github.com/jeffbyrnes/e56d294c216fbd30fd2fd32e576db81c)


### Generating Weights

The weights are used to interpolate DWD ICON Data from an icosahedral grid to a regular grid. You can generate the weights with the following commands:
```bash
chmod +x generateWeights.sh
./generateWeights.sh
```

This will download about 1.2GB of data and generate the weights.
The computed weights are stored in the `weights` directory.

### Running:
You can run gribdl with the following command without compiling it into a standalone binary:
```bash
cd src && go run main.go --help
```

### Building:
You can also build gribdl into a standalone binary with the following command:
```bash
cd src && go build main.go -o gribdl
```
