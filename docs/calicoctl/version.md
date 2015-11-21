> You are viewing the calico-docker documentation for release v0.12.0.

# User reference for 'calicoctl version' commands

This sections describes the `calicoctl version` commands.

This command prints the version of `calicoctl` in use.

Read the [calicoctl command line interface user reference](../calicoctl.md) 
for a full list of calicoctl commands.

## Displaying the help text for 'calicoctl version' commands

Run `calicoctl version --help` to display the following help menu for the 
calicoctl version commands.

```

Usage:
  calicoctl version

Description:
  Display the version of calicoctl

```

## calicoctl version commands


### calicoctl version

Print the version of `calicoctl` in use.

This command is specific to the `calicoctl` being run on a given machine.

Command syntax:

```
calicoctl version

```

Examples:

```
$ calicoctl version
0.8.0
```