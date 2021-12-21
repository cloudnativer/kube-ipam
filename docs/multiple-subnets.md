# multiple subnets

`ranges` can support the configuration format of multiple subnets.


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
                    "subnet": "10.199.0.0/16"
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
                "subnet": "10.166.0.0/16",
                "rangeStart": "10.166.0.30",
                "rangeEnd": "10.166.0.200"
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


