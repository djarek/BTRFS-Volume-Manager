# BTRFS-Volume-Manager [![Build Status](https://travis-ci.org/djarek/btrfs-volume-manager.svg?branch=master)](https://travis-ci.org/djarek/btrfs-volume-manager)
## A responsive and easy to use BTRFS based NAS server management solution

## Business justification
BTRFS (B-tree file system) is a copy on write (CoW) filesystem for Linux aimed at implementing advanced features while focusing on fault tolerance, repair and easy administration. It is a relatively novel filesystem so we decided to try develop a tool for managing volumes so it can be useful in many ways of usage.
## State of Art (what are the competitive solutions on the market, what has been achieve so far in that project area on the market)
There are several projects on market which are providing tools for managing BTRFS like [RockStor](http://rockstor.com/) or [FreeNAS](http://www.freenas.org/). 
## Project scope 
#### Goal / Aims 
  - Provide an easy to use interface to manage devices, BTRFS filesystems and snapshots
  - The target user is a sysadmin managing a number of NAS  devices
  - The solution should be accessible via web technologies
  - The system has to be secure enough to be useful in enterprise environments

#### Core features
  - Presentation of basic information about each BTRFS volume
  - Creation and deletion of snapshots & subvolumes
  - Provide the user with basic statistics
  - Allow management of multiple BTRFS-based servers with one utility
  - Allow scheduling of some tasks (e.g. snapshot a certain subvolume every few hours)
  - Provide the user with an easy-to-use SPA type GUI

#### Optional features
  - Allow viewing of S.M.A.R.T information for each block device on any of the attached servers
  - Management of RAID functionality of volumes
  - Provide an API for easy creation of scripts to automate some tasks (e.g. create a snapshot whenever a certain process starts or stops)
  - Scheduler for handling of incremental backups (through btrfs send/recv)
  - Provide utility for setting up Samba and NFS shares

#### Project's risks
  - New programming language for programmers
  - New technologies
  - Complex topic
  - Lack of enough time

## Techniques & Technologies used to implement your project
  - [Go](https://golang.org/doc/) - backend programming language
  - [HTML](http://devdocs.io/html/) - client-side markup language
  - [AngularJS](https://docs.angularjs.org/api) - client-side web application language
  - [C](http://en.cppreference.com/w/c) - low-level communication with the Linux kernel)
  - [MongoDB](https://docs.mongodb.com/) - NoSQL database

## Milestones & Project plan
#### Gantt Chart
![Gantt Chart](https://i.imgsafe.org/6b7e858e1b.png)

#### Cost & effort calculations 
  - Work at home: 3 people x 80 hours = 240 hours
  - Payment:  320h x 60 = 19200
  - Social security and insurance (19,64%): 3771
  - Indirect costs: 0
  - VAT (23%): 5284
  - Product cost: 28254

#### System architecture
![System architecture](https://i.imgsafe.org/6b8d060ca5.png)

## Thorough Implementation description along with deployment information
#### Steps to install

#### Hardware & Software requirements

## Conclusions
