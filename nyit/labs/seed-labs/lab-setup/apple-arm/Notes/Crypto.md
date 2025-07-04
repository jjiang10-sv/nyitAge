# Crypto Labs

## Status update

1/31/2024: Added the corresponding setup files to each of the 
labs. The new setup files are put inside `Labsetup-arm`. All
the required docker images (arm version) are already pushed 
to Docker Hub. The lab descriptions do not need to change. 

TODO:  update the lab manual for the MD5 lab. No need to revise
the manuals for the other labs. 


## Secret-Key Encryption Lab

The openssl library is compiled successfully. 
Remember to run `openssl.sh` when building the VM.


## Padding Oracle Attack Lab

We have pushed the docker image to Docker Hub. 
We followed the instructions in 
`seed-labs/category-crypto/Crypto_Padding_Oracle/Container_Building`
regarding how to create the container. 


## MD5 Collision Attack

We need to rebuild the `md5collgen` program from the source. 
The one included in the `Labsetup.zip` file is the AMD version. 
Using the shell script `md5_firsttime.sh` in 
`seed-labs/lab-setup/ubuntu20.04-vm/src-vm`, we can successfully
build the `md5collgen` program. 

We also need to modify the instructor manuals for this 
lab: the arm binary and amd binary are different, so the offsets
of the array in the binary are also different. The offsets
are used in the solution script.


## Hash Length Extension

The Flask version has an issue. We can add the following line
to the `Dockerfile` to fix the issue (similar to the DNS rebinding 
lab). 

```
RUN pip3 install flask --upgrade
```

In the Flask docker image, we installed an old version using the 
following. We may need to update this, but we need to make sure 
it does not break other labs. 

```
pip3 install flask==1.1.1
```

## Random Number

The way how `/dev/random` works on Apple silicon machines is quite different.
I believe that most of the tasks can be made to work, but we need to modify the 
lab description. 

Updated on 3/27/2024: It seems that the behavior of the `/dev/random` has 
changed in the recent Linux kernel. The behaviors for Ubuntu 20.04 and 22.04 
are different (the ARM VM uses 22.04). We filed an issue (#141) regarding 
this. The lab needs to be redesigned. This is not an issue with the ARM 
machines, but an issue with the Linux version.  
