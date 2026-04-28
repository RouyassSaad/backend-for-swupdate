/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

#include "cgo.h"

// thats odd, but we have to explicitly call every union members that we want to work with in go
// we implement for every impl_union_utils(<foo>, <bar>):
// 1. a function: cast_union_<foo>_as_member_<bar>
// 2. a function: cast_member_<bar>_as_union_<foo>
impl_union_utils(msgdata, msg);
impl_union_utils(msgdata, status);
impl_union_utils(msgdata, notify);
impl_union_utils(msgdata, instmsg);
impl_union_utils(msgdata, procmsg);
impl_union_utils(msgdata, aeskeymsg);
impl_union_utils(msgdata, versions);
impl_union_utils(msgdata, revisions);
impl_union_utils(msgdata, vars);


impl_union_utils(wrapper_ipc_message, msg);
impl_union_utils(wrapper_progress_msg, msg);
