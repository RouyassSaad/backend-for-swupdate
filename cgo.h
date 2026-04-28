/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

#pragma once

// utility to interact with C union in Go
#include "cgo_union.h"

#include <network_ipc.h>
#include <progress_ipc.h>

// thats odd, but we have to explicitly call every union members that we want to work with in go
// we define for every def_union_utils(<foo>, <bar>):
// 1. a type: union_<foo>_as_member_<bar>
// 2. a function: cast_union_<foo>_as_member_<bar>
// 3. a function: cast_member_<bar>_as_union_<foo>
def_union_utils(msgdata, msg);
def_union_utils(msgdata, status);
def_union_utils(msgdata, notify);
def_union_utils(msgdata, instmsg);
def_union_utils(msgdata, procmsg);
def_union_utils(msgdata, aeskeymsg);
def_union_utils(msgdata, versions);
def_union_utils(msgdata, revisions);
def_union_utils(msgdata, vars);


// union wrapper around types we want to send by ipc as buffer
typedef union {
	ipc_message msg;
} wrapper_ipc_message;

def_union_utils(wrapper_ipc_message, msg);

typedef union {
	struct progress_msg msg;
} wrapper_progress_msg;

def_union_utils(wrapper_progress_msg, msg);

// Cgo does not really understand that the type RECOVERY_STATUS from <swupdate_status.h>
// can be used as progress_msg.status, lets make a wrapper to it
typedef __typeof__(((struct progress_msg*)0)->status) C_RECOVERY_STATUS;

