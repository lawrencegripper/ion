# Demo: Create a pipeline which downloads and transcodes video

## Pre-reqs

1. Have an existing ion cluster
2. Have your `~/.ion.yaml` file configuration with the connection details
3. Have the `ion` cli in your `PATH`

# 1. Build and Deploy [Downloader module](./../modules/downloader)

Lets use docker to build a module and publish it to our docker repository. 

``` bash

docker build -t lawrencegripper/ion-module-downloader -f ./modules/downloader/Dockerfile .
docker push lawrencegripper/ion-module-downloader 

```

Now we can use the ION cli to create a module in our pipeline from this docker image. The `-i` and `-o` arguments represent the `input` and `output` events. 
The `frontapi.new_link` event in outputted by the [FrontAPI used to submit items to the pipeline](./../cmd/frontapi). The `file_downloaded` event is what we'll output
after we've downloaded the file. Lastly the `-p` specifies the provider to run the module under. In this case we'll use a shared Kuberentes environment and it will run as a [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/). 

``` bash 

ion module create --module-image lawrencegripper/ion-module-downloader -i frontapi.new_link -o file_downloaded -p Kuberentes

```

> *Note*: update `lawrencegripper/` to your dockerhub username

# 2. Build and Deploy the [Transcoder module](./../modules/transcoder)

Now lets build the transcoder module, same as before we'll make and upload a docker image. 

``` bash 

docker build -t lawrencegripper/ion-module-transcode -f ./modules/transcode/Dockerfile .
docker push lawrencegripper/ion-module-transcode 

```

Now again we'll use the ION cli to create the module but this time we'll subscribe to the `file_downloaded` event we just outputted from the downloader module and we'll use `-p AzureBatch` to schedule this work on an [AzureBatch pool](https://docs.microsoft.com/en-us/azure/batch/) with GPU support so we can use the GPU for transcoding. 

``` bash 

ion module create --module-image lawrencegripper/ion-module-transcode -i file_downloaded -o file_transcoded -p AzureBatch

```

# 3. Submit a link to the pipeline 


So next lets trigger our pipeline with a new URL to download and transcode. We'll use CURL to POST [out request](./data/examplesubmission.json) for the [Big Buck Bunny film](https://peach.blender.org/).

``` bash 

curl --header "Content-Type: application/json"   --request POST   --data-binary "./docs/data/examplesubmission.json"   http://youendpointhere:9001/

```

This will give you a response with some useful information, something like this:

``` json 

{
    "context": {
        "name": "frontapi",
        "eventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88",
        "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
        "parentEventId": "frontapi",
        "documentType": "eventMeta"
    },
    "type": "frontapi.new_link"
}

```

# 4. Debug and get results

In the response from your link submission you'll be returned a `correlationId` this lets you track all work associated with this submission through the pipeline using the ION cli.

``` bash 

ion trace flow -c your-correlation-id-here

```

> *NOTE*: Update the command and put your `correlationId` in place of `your-correlation-id-here` for example `0bff834f-4dfc-484a-80ed-4f2cfc149c23` from the response above. 

Congratulations, you've run a pipeline in ION. You'll now get information back about the execution of the pipeline. 

Here is a quick guide to reading this output. 

- `context.documentType` can be either:
    - `eventMeta`: This is data transferred between events used by the system. Then include links to blob data uploaded by a module so you can debug the flow between modules. 
    - `insight`: This is data stored by the module for querying by the user, this could include execution time, items spotted in the video or any arbitrary data for query. You can inspect the putput here. 
    - `modulelogs`: This contains a link to the console logs (`stdout/stderr`) outputted by the module when it ran. These documents contain a link to the full log.

- The `context` object more generally gives context of the module which ran and which event triggered it, `parentEventId`. 

``` json 

[
  {
    "_id": "5bb218f98639072364ffc062",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "eventMeta",
      "eventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88",
      "name": "frontapi",
      "parentEventId": "frontapi"
    },
    "data": [
      {
        "key": "url",
        "value": "http://download.blender.org/peach/bigbuckbunny_movies/BigBuckBunny_320x180.mp4"
      }
    ],
    "files": [],
    "id": "b9f9d945-3f0d-4650-9c11-58bd761e8a88"
  },
  {
    "_id": "5bb21909d87fb45c5c3529b2",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "insight",
      "eventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88",
      "name": "downloader",
      "parentEventId": "frontapi"
    },
    "data": [
      {
        "key": "downloadTimeSec",
        "value": "3.247324"
      },
      {
        "key": "sourceUrl",
        "value": "http://download.blender.org/peach/bigbuckbunny_movies/BigBuckBunny_320x180.mp4"
      }
    ],
    "id": "8d23c311-863f-44e2-8e95-aafe413fe4bb"
  },
  {
    "_id": "5bb21909d87fb45c5c3529b3",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "eventMeta",
      "eventId": "fce68eb3-fca4-4a54-85b0-7eea74252e4f",
      "name": "downloader",
      "parentEventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88"
    },
    "data": [
      {
        "key": "file.raw",
        "value": "https://yourstuffhere.blob.core.windows.net/0bff834f-4dfc-484a-80ed-4f2cfc149c23/b9f9d945-3f0d-4650-9c11-58bd761e8a88/downloader/ion/file.raw?se=2018-10-02T12%3A54%3A33Z&sig=TxeFudgko5LUMpm5MeAGF01qf%2BSqzvNa4L7HI3CYG2s%3D&sp=r&spr=https&sr=b&st=2018-10-01T11%3A54%3A33Z&sv=2015-04-05"
      }
    ],
    "files": [
      "file.raw"
    ],
    "id": "fce68eb3-fca4-4a54-85b0-7eea74252e4f"
  },
  {
    "_id": "5bb2190f8639072364ffc063",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "modulelogs",
      "eventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88",
      "name": "downloader",
      "parentEventId": "frontapi"
    },
    "desc": "module:downloader-event:b9f9d945-3f0d-4650-9c11-58bd761e8a88-attempt:0",
    "id": "module:downloader-event:b9f9d945-3f0d-4650-9c11-58bd761e8a88-attempt:0",
    "logs": "https://yourstuffhere.blob.core.windows.net/logs/0bff834f-4dfc-484a-80ed-4f2cfc149c23/b9f9d945-3f0d-4650-9c11-58bd761e8a88/frontapi-attempt-0.log?se=2018-10-02T12%3A54%3A39Z&sig=EQwCLZd4uTv%2F%2Fjlw5pw8nNtERSkEomP1cJBEfyhT0Dg%3D&sp=r&sr=b&st=2018-10-01T11%3A54%3A39Z&sv=2016-05-31",
    "succeeded": true
  },
  {
    "_id": "5bb2196367c2a45bf8542bc1",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "insight",
      "eventId": "fce68eb3-fca4-4a54-85b0-7eea74252e4f",
      "name": "transcode",
      "parentEventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88"
    },
    "data": [
      {
        "key": "transcodeTimeSec",
        "value": "15.213627"
      }
    ],
    "id": "bb548f42-19ab-44f4-bd73-a5ed4cf808b9"
  },
  {
    "_id": "5bb2196367c2a45bf8542bc2",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "eventMeta",
      "eventId": "ffa1e284-4d70-4b4e-bbcc-777e1be68962",
      "name": "transcode",
      "parentEventId": "fce68eb3-fca4-4a54-85b0-7eea74252e4f"
    },
    "data": [
      {
        "key": "file.raw-1280x720-h264.mp4",
        "value": "https://yourstuffhere.blob.core.windows.net/0bff834f-4dfc-484a-80ed-4f2cfc149c23/fce68eb3-fca4-4a54-85b0-7eea74252e4f/transcode/ion/file.raw-1280x720-h264.mp4?se=2018-10-02T12%3A56%3A03Z&sig=ao%2Bue8RhPITcWIsYw7ioB%2BH7sVFxjz8BzKgO5totRp8%3D&sp=r&spr=https&sr=b&st=2018-10-01T11%3A56%3A03Z&sv=2015-04-05"
      }
    ],
    "files": [
      "file.raw-1280x720-h264.mp4"
    ],
    "id": "ffa1e284-4d70-4b4e-bbcc-777e1be68962"
  },
  {
    "_id": "5bb2196b7ff4153294a7d08d",
    "context": {
      "correlationId": "0bff834f-4dfc-484a-80ed-4f2cfc149c23",
      "documentType": "modulelogs",
      "eventId": "fce68eb3-fca4-4a54-85b0-7eea74252e4f",
      "name": "transcode",
      "parentEventId": "b9f9d945-3f0d-4650-9c11-58bd761e8a88"
    },
    "desc": "module:transcode-event:fce68eb3-fca4-4a54-85b0-7eea74252e4f-attempt:0",
    "id": "module:transcode-event:fce68eb3-fca4-4a54-85b0-7eea74252e4f-attempt:0",
    "logs": "https://yourstuffhere.blob.core.windows.net/logs/0bff834f-4dfc-484a-80ed-4f2cfc149c23/fce68eb3-fca4-4a54-85b0-7eea74252e4f/downloader-attempt-0.log?se=2018-10-02T12%3A56%3A11Z&sig=nXgaU8BbO0SpFn5xBB8H%2F4mvDkKLgV%2BpiUHQvjJXT80%3D&sp=r&sr=b&st=2018-10-01T11%3A56%3A11Z&sv=2016-05-31",
    "succeeded": true
  }
]


```