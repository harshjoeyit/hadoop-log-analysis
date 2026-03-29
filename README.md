## Hadoop MapReduce prototype

### Architecture

#### DFS = Distributed File System
         +------------------+
         |   NameNode       |
         | (Metadata brain) |
         +--------+---------+
                  |
                  |
         +--------v---------+
         |   DataNode       |
         | (Stores data)    |
         +------------------+
NameNode → knows where data lives
DataNode → actually stores data blocks

- Namenode is synonymous with Master server in GFS
- Datanode is synonymous with Chunk server in GFS

```
DFS (HDFS)
│
├── NameNode (metadata)
│     └── /hadoop/dfs/name
│
└── DataNode (actual data)
      └── /hadoop/dfs/data
```

#### [Google file system (GFS)](https://www.notion.so/harshjoeyit/Short-notes-based-on-video-2194246633a38084a4f0edaf2303e1cb)


## The Scenario
You're analyzing NASA's web server logs (real public dataset) to answer:

- Which pages get most traffic?
- What are peak usage hours?
- Which countries generate most 404 errors?
- Identify potential bot/scraper activity

```
# NASA HTTP logs (190MB - good starter size)
wget ftp://ita.ee.lbl.gov/traces/NASA_access_log_Jul95.gz
wget ftp://ita.ee.lbl.gov/traces/NASA_access_log_Aug95.gz
```

**Sample log line:**
```
199.72.81.55 - - [01/Jul/1995:00:00:01 -0400] "GET /history/apollo/ HTTP/1.0" 200 6245
```

## Running locally

```
docker-compose  up -d 
```
```
❯ docker ps
CONTAINER ID   IMAGE                                             COMMAND                  CREATED          STATUS                    PORTS                                         NAMES
fd3470bfed4f   bde2020/hadoop-datanode:2.0.0-hadoop3.2.1-java8   "/entrypoint.sh /run…"   48 seconds ago   Up 47 seconds (healthy)   9864/tcp                                      datanode
1f97612348b3   bde2020/hadoop-namenode:2.0.0-hadoop3.2.1-java8   "/entrypoint.sh /run…"   49 seconds ago   Up 48 seconds (healthy)   0.0.0.0:9870->9870/tcp, [::]:9870->9870/tcp   namenode
```
```
# stop containers
docker-compose down

# stop containers, remove volumes and remove any orphan containers
docker-compose down -v --remove-orphans
```


#### Uploading files to HDFS
```
❯ curl -O ftp://ita.ee.lbl.gov/traces/NASA_access_log_Jul95.gz
❯ curl -O ftp://ita.ee.lbl.gov/traces/NASA_access_log_Aug95.gz
```

```
❯ gunzip NASA_access_log_Jul95.gz
❯ gunzip NASA_access_log_Aug95.gz
```

```
❯ docker cp NASA_access_log_Aug95 namenode:/tmp/
Successfully copied 168MB to namenode:/tmp/
❯ docker cp NASA_access_log_Jul95 namenode:/tmp/
Successfully copied 205MB to namenode:/tmp/
```

```
❯ docker exec -it namenode bash
root@9d490d763199:/# ls -ltr /tmp
total 364336
-rw-r--r-- 1  501 dialout 205242368 Mar 28 06:49 NASA_access_log_Jul95
-rw-r--r-- 1  501 dialout 167813770 Mar 28 07:03 NASA_access_log_Aug95
-rw-r--r-- 1 root root            4 Mar 28 07:11 hadoop-root-namenode.pid
drwxr-xr-x 1 root root         4096 Mar 28 07:11 hsperfdata_root
```

```
root@9d490d763199:/# hdfs dfs -mkdir -p /data/nasa
```

```
root@9d490d763199:/# hdfs dfs -put /tmp/NASA_access_log_Jul95 /data/nasa/
2026-03-28 07:16:05,284 INFO sasl.SaslDataTransferClient: SASL encryption trust check: localHostTrusted = false, remoteHostTrusted = false
2026-03-28 07:16:08,873 INFO sasl.SaslDataTransferClient: SASL encryption trust check: localHostTrusted = false, remoteHostTrusted = false

root@9d490d763199:/# hdfs dfs -put /tmp/NASA_access_log_Aug95 /data/nasa/
2026-03-28 07:16:20,112 INFO sasl.SaslDataTransferClient: SASL encryption trust check: localHostTrusted = false, remoteHostTrusted = false
2026-03-28 07:16:22,324 INFO sasl.SaslDataTransferClient: SASL encryption trust check: localHostTrusted = false, remoteHostTrusted = false
root@9d490d763199:/# 
```
### Replication

Note: 0. and 1. denote each individual blocks being replicated thrice.
```
Block 0: DN5, DN4, DN3
Block 1: DN4, DN5, DN3
```
```
root@9d490d763199:/# hdfs fsck /data/nasa/NASA_access_log_Jul95 -files -blocks -locations

Connecting to namenode via http://namenode:9870/fsck?ugi=root&files=1&blocks=1&locations=1&path=%2Fdata%2Fnasa%2FNASA_access_log_Jul95

FSCK started by root (auth:SIMPLE) from /172.19.0.2 for path /data/nasa/NASA_access_log_Jul95 at Sat Mar 28 09:37:07 UTC 2026
/data/nasa/NASA_access_log_Jul95 205242368 bytes, replicated: replication=3, 2 block(s):  OK

0. BP-1750460713-172.19.0.2-1774681868114:blk_1073741825_1001 len=134217728 Live_repl=3  [
      DatanodeInfoWithStorage[172.19.0.5:9866,DS-84d82a00-8fea-4796-847a-8ff2579702ee,DISK], 
      DatanodeInfoWithStorage[172.19.0.4:9866,DS-9532e936-81e0-4ab9-ba56-5fb98816dc5d,DISK], 
      DatanodeInfoWithStorage[172.19.0.3:9866,DS-4ba9b105-ec7c-4ef7-9d81-57d06fbec8f0,DISK]
      ]
1. BP-1750460713-172.19.0.2-1774681868114:blk_1073741826_1002 len=71024640 Live_repl=3  [
      DatanodeInfoWithStorage[172.19.0.4:9866,DS-9532e936-81e0-4ab9-ba56-5fb98816dc5d,DISK], 
      DatanodeInfoWithStorage[172.19.0.5:9866,DS-84d82a00-8fea-4796-847a-8ff2579702ee,DISK], 
      DatanodeInfoWithStorage[172.19.0.3:9866,DS-4ba9b105-ec7c-4ef7-9d81-57d06fbec8f0,DISK]
      ]
```

### MapReduce

MapReduce fundamentally = stdin → stdout transformation

#### 🧩 Mental model
```
Mapper = emit (key, value)
Reducer = aggregate by key
```
Example of reduce function in javascript to build understanding

```
var pets = ['dog', 'chicken', 'cat', 'dog', 'chicken', 'chicken', 'rabbit'];

var petCounts = pets.reduce(function(obj, pet){
    if (!obj[pet]) {
        obj[pet] = 1;
    } else {
        obj[pet]++;
    }
    return obj;
}, {});

console.log(petCounts); 

/*
Output:
 { 
    dog: 2, 
    chicken: 3, 
    cat: 1, 
    rabbit: 1 
 }
 */
```

#### 💡 Where does MapReduce runs? / What are the compute nodes?

In Hadoop, compute runs on the same machines that store data i.e. Datanodes

Step-by-step flow
1. Job submitted (from NameNode container)
2. Hadoop splits input file into blocks
3. Each block assigned to a node that already has it
4. Mapper runs ON that node

#### Testing mapper and reducer (locally, not on HDFS)

Mapper
```
❯ echo '127.0.0.1 - - [10/Oct/2023:13:55:36 -0700] "GET /api/v1/users HTTP/1.1" 200 1234' | go run mapper/main.go
/api/v1/users   1
```

Reducer
```
❯ echo -e "abc\t1\ncdb\t1\nabc\t1" | go run reducer/main.go
abc     2
cdb     1
```

#### Generate binaries

```
❯ GOOS=linux GOARCH=amd64 go build -o mymapper ./mapper
❯ GOOS=linux GOARCH=amd64 go build -o reducer ./reducer
```

#### ERROR: Architecture mismatch MacOS vs Linux
```
T: not found
/tmp/mymapper: 15: /tmp/mymapper: 
                                  ?@9b?@?d@9#6D@9?7?/??C?@????@??+?+?@?8?!??K???@??+@?H?@?L*????D????C@??/@????C9?????_??I??_?B?_?J??Thb8?qa??TO??!?????Tha8??qa??Ta????: not found
/tmp/mymapper: 1: /tmp/mymapper: Syntax error: Unterminated quoted string
2026-03-28 10:54:26,622 WARN streaming.PipeMapRed: {}
```

#### Copy to binaries to docker container 

```
❯ docker cp mymapper namenode:/tmp/
Successfully copied 2.45MB to namenode:/tmp/
❯ docker cp myreducer namenode:/tmp/
Successfully copied 2.39MB to namenode:/tmp/
```

#### Make executable

```
❯ docker exec -it namenode bash
root@9d490d763199:/# chmod +x /tmp/mymapper /tmp/myreducer
```

### Run MapReduce via streaming

```
hadoop jar $HADOOP_HOME/share/hadoop/tools/lib/hadoop-streaming-*.jar \
  -input /data/nasa/NASA_access_log_Jul95 \
  -output /output/url_counts \
  -mapper /tmp/mymapper \
  -reducer /tmp/myreducer

```


#### To rerun
Delete output file
```
❯ docker exec -it namenode bash
root@9d490d763199:/# hdfs dfs -rm -r /output/url_counts
Deleted /output/url_counts
```

#### Output logs

```
ile System Counters
                ...
                ... 
                HDFS: Number of bytes read=544714752
                HDFS: Number of bytes written=804153
                ...
                ...
        Map-Reduce Framework
                ...
                ...
                Map input records=1891715
                Map output records=1884653
                Map output bytes=64446142
                ...
                ...
                Reduce input groups=21616
                Reduce input records=1884653
                Reduce output records=21615
                ...
                ...
                Shuffled Maps =2
                Failed Shuffles=0
                Merged Map outputs=2
```

### Result: Top 20 most visited URLs
```
root@9d490d763199:/# hdfs dfs -cat /output/url_counts/part-00000 | sort -t$'\t' -k2 -rn | head -20
2026-03-29 07:01:14,582 INFO sasl.SaslDataTransferClient: SASL encryption trust check: localHostTrusted = false, remoteHostTrusted = false
/images/NASA-logosmall.gif      110681
/images/KSC-logosmall.gif       89357
/images/MOSAIC-logosmall.gif    59969
/images/USA-logosmall.gif       59516
/images/WORLD-logosmall.gif     58999
/images/ksclogo-medium.gif      58413
/images/launch-logo.gif 40780
/shuttle/countdown/     40143
/ksc.html       39854
/images/ksclogosmall.gif        33528
/       32535
/history/apollo/images/apollo-logo1.gif 31011
/shuttle/missions/missions.html 24745
/htbin/cdt_main.pl      22572
/shuttle/countdown/count.gif    22167
/shuttle/countdown/liftoff.html 21946
/shuttle/countdown/count70.gif  20859
/images/launchmedium.gif        20753
/shuttle/missions/sts-71/sts-71-patch-small.gif 19810
/shuttle/missions/sts-70/sts-70-patch-small.gif 18107
```

### Todo:
1. Failure of data node
2. Consider adding Apache Spark on top
3. Strengthen Hadoop fundamentals. → add YARN + deeper MapReduce
4. Move to modern stack. → Spark + real analytics pipeline