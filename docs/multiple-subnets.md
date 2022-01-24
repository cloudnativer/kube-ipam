# multiple subnets

`ranges` of `/etc/cni/net.d/1-kube-ipam.conf` can support the configuration format of multiple subnets.
<br>
If you have multiple subnets to configure, you can insert multiple `range` contents. 
<br>
Here are some examples for your reference.


## Example 1:

```
"ranges": [
            [
                {
                    "subnet": "10.188.0.0/16",
                    "fixedStart": "10.188.0.10",
                    "fixedEnd": "10.188.3.155",
                    "rangeStart": "10.188.3.156",
                    "rangeEnd": "10.188.255.254",
                    "gateway": "10.188.0.1"
                },
                {
                    "subnet": "10.199.0.0/16",
                    "fixedStart": "10.199.0.10",
                    "fixedEnd": "10.199.0.255",
                    "rangeStart": "10.199.1.0",
                    "rangeEnd": "10.199.255.254",
                    "gateway": "10.199.0.1"
                }
            ]
        ]
```


## Example 2:

```
"ranges": [
            [{
                    "subnet": "10.188.0.0/16",
                    "fixedStart": "10.188.0.10",
                    "fixedEnd": "10.188.0.255",
                    "rangeStart": "10.188.1.0",
                    "rangeEnd": "10.188.255.254",
                    "gateway": "10.188.0.1"
            }],
            [{
                    "subnet": "10.166.0.0/16"
            }]
        ]
```



## Example 3:

```
"ranges": [
            [
                {
                    "subnet": "10.188.0.0/16",
                    "rangeStart": "10.188.0.150",
                    "rangeEnd": "10.188.255.254",
                    "gateway": "10.188.0.1"
                },
                {
                    "subnet": "10.199.0.0/16"
                }
            ],
            [{
                  "subnet": "10.166.0.0/16",
                  "fixedStart": "10.166.0.10",
                  "fixedEnd": "10.166.0.200",
                  "rangeStart": "10.166.0.201",
                  "rangeEnd": "10.166.255.254"
            }]
        ]
```


