/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

#pragma once

#define _inner_concat(a, b) a##b
#define _concat(a, b) _inner_concat(a, b)

// since cgo is not able to understand correctly unions, lets create macros to help us on it


// for an union foo { mytype bar; }, 
// union_as_member_type_name(foo, bar) is union_foo_as_member_bar
// union_as_member_func_name(foo, bar) is cast_union_foo_as_member_bar
// member_as_union_func_name(foo, bar) is cast_member_bar_as_union_foo
#define union_as_member_type_name(u, t) union_##u##_as_member_##t
#define union_as_member_func_name(u, t) _concat(cast_, union_as_member_type_name(u, t))
#define member_as_union_func_name(u, t) cast_member_##t##_as_union_##u


// typedef the member type by infering it: typedef typeof(foo_instance->bar) union_foo_as_member_bar
#define def_type_union_as_member(u, t) typedef __typeof__(((u*)0)->t) union_as_member_type_name(u, t)


// create a function to "cast" the union type to member type:
// union_foo_as_member_bar * cast_union_foo_as_member_bar(foo * input) { return &(input->bar); }
#define def_cast_union_as_member(u, t) \
union_as_member_type_name(u, t) * union_as_member_func_name(u, t) (u * input)

#define impl_cast_union_as_member(u, t) \
def_cast_union_as_member(u, t) { return &(input->t); }


// def_* macros should be used in a .h file 
#define def_union_utils(u, t) \
	def_type_union_as_member(u, t); \
	def_cast_union_as_member(u, t);

// impl_* macros should be used in a .c file
#define impl_union_utils(u, t) \
	impl_cast_union_as_member(u, t);

