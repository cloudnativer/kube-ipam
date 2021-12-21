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
                    "rangeStart": "10.188.0.10",
                    "rangeEnd": "10.188.0.100",
                    "gateway": "10.188.0.1"
                },
                {
                    "subnet": "10.199.0.0/16",
                    "rangeStart": "10.199.0.10",
                    "rangeEnd": "10.199.0.100",
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
                    "rangeStart": "10.188.0.10",
                    "rangeEnd": "10.188.0.100",
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
                    "rangeStart": "10.188.0.10",
                    "rangeEnd": "10.188.0.100",
                    "gateway": "10.188.0.1"
                },
                {
                    "subnet": "10.199.0.0/16"
                }
            ],
            [{
                  "subnet": "10.166.0.0/16",
                  "rangeStart": "10.166.0.30",
                  "rangeEnd": "10.166.0.200"
            }]
        ]
```


