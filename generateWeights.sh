#!/bin/bash


mkdir -p /tmp/gribdl/dwd/grids
mkdir -p ./weights

function downloadGrids {
    if [ -f /tmp/gribdl/dwd/grids/${1%.bz2} ]; then
        echo "File ${1%.bz2} already exists"
        return
    fi
    echo "Downloading $1"
    wget -O - https://opendata.dwd.de/weather/lib/cdo/$1 | bunzip2 > /tmp/gribdl/dwd/grids/${1%.bz2}
}


downloadGrids icon_grid_0047_R19B07_L.nc.bz2 &
downloadGrids icon_grid_0028_R02B07_N02.nc.bz2 &
downloadGrids icon_grid_0024_R02B06_G.nc.bz2 &
downloadGrids icon_grid_0026_R03B07_G.nc.bz2 &

wait

echo "All grids downloaded"

# ICON-D2
cdo gennn,./weights/icon-d2_description.txt \
    -setgrid,/tmp/gribdl/dwd/grids/icon_grid_0047_R19B07_L.nc:2 \
    ./weights/icon-d2_sample.grib2 \
    ./weights/icon-d2_weights.nc

# ICON-D2-EPS
cdo gennn,./weights/icon-d2-eps_description.txt \
    -setgrid,/tmp/gribdl/dwd/grids/icon_grid_0047_R19B07_L.nc:2 \
    ./weights/icon-d2-eps_sample.grib2 \
    ./weights/icon-d2-eps_weights.nc

# ICON-EU-EPS
cdo gennn,./weights/icon-eu-eps_description.txt \
    -setgrid,/tmp/gribdl/dwd/grids/icon_grid_0028_R02B07_N02.nc:1 \
    ./weights/icon-eu-eps_sample.grib2 \
    ./weights/icon-eu-eps_weights.nc

# ICON-EU
cdo gennn,./weights/icon-eps_description.txt \
    -setgrid,/tmp/gribdl/dwd/grids/icon_grid_0024_R02B06_G.nc:1 \
    ./weights/icon-eps_sample.grib2 \
    ./weights/icon-eps_weights.nc

# ICON
cdo gennn,./weights/icon_description.txt \
    -setgrid,/tmp/gribdl/dwd/grids/icon_grid_0026_R03B07_G.nc:1 \
    ./weights/icon_sample.grib2 \
    ./weights/icon_weights.nc