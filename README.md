# SSH LXD

A proof of concept for an ssh server that spawns a bash session inside a LXD container.

## TODO

- [ ] Auth check
- [ ] Multiuser support
- [ ] ???


## Auth

Currently work in progress.

Authorized keys can be provided by creating a config.yaml with the name of the Instances and a keys array

```yaml
auth:
  "poetic-crow":
    keys:
      - "< some ssh key >"
      - "< some ssh key >"
  "trusty-titmouse":
    keys:
      - "< some ssh key >"
```
