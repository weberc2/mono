# DSM Spaces

## Map Tiles

* Generate ./data/des-moines.pmtiles file via:

   ```bash
   $ docker run -e JAVA_TOOL_OPTIONS="-Xmx6g" \
       -v "$(pwd)/data":/data \
       ghcr.io/onthegomap/planetiler:latest \
       --download \
       --area=us/iowa \
       --bounds=-93.75,41.48,-93.45,41.72 \
       --output=/data/des-moines.pmtiles --force
   ```

* This is served to the frontend by the `/tiles/desmoines.pmtiles` backend route

* This must exist before the application is run or else it will fail