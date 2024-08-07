/* -*- Mode: C; c-basic-offset:4 ; indent-tabs-mode:nil -*- */
/*
 * Copyright (c) 2010      Cisco Systems, Inc.  All rights reserved.
 * Copyright (c) 2004-2011 The University of Tennessee and The University
 *                         of Tennessee Research Foundation.  All rights
 *                         reserved.
 * Copyright (c) 2011-2014 Los Alamos National Security, LLC.  All rights
 *                         reserved.
 * Copyright (c) 2014-2015 Intel, Inc. All rights reserved.
 * Copyright (c) 2015      Research Organization for Information Science
 *                         and Technology (RIST). All rights reserved.
 * $COPYRIGHT$
 *
 * Additional copyrights may follow
 *
 * $HEADER$
 *
 */

#include <private/autogen/config.h>
#include <pmix/rename.h>
#include <private/pmix_stdint.h>
#include <private/hash_string.h>

#include <string.h>

#include "src/include/pmix_globals.h"
#include "src/class/pmix_hash_table.h"
#include "src/class/pmix_pointer_array.h"
#include "src/buffer_ops/buffer_ops.h"
#include "src/util/error.h"
#include "src/util/output.h"

#include "src/util/hash.h"

/**
 * Data for a particular pmix process
 * The name association is maintained in the
 * proc_data hash table.
 */
typedef struct {
    /** Structure can be put on lists (including in hash tables) */
    pmix_list_item_t super;
    /* List of pmix_kval_t structures containing all data
       received from this process */
    pmix_list_t data;
} pmix_proc_data_t;
static void pdcon(pmix_proc_data_t *p)
{
    PMIX_CONSTRUCT(&p->data, pmix_list_t);
}
static void pddes(pmix_proc_data_t *p)
{
    PMIX_LIST_DESTRUCT(&p->data);
}
static PMIX_CLASS_INSTANCE(pmix_proc_data_t,
                           pmix_list_item_t,
                           pdcon, pddes);

static pmix_kval_t* lookup_keyval(pmix_list_t *data,
                                  const char *key);
static pmix_proc_data_t* lookup_proc(pmix_hash_table_t *jtable,
                                     uint64_t id, bool create);

int pmix_hash_store(pmix_hash_table_t *table,
                    int rank, pmix_kval_t *kin)
{
    pmix_proc_data_t *proc_data;
    uint64_t id;

    pmix_output_verbose(10, pmix_globals.debug_output,
                        "HASH:STORE rank %d key %s",
                        rank, kin->key);

    id = (uint64_t)rank;

    /* lookup the proc data object for this proc - create
     * it if we don't already have it */
    if (NULL == (proc_data = lookup_proc(table, id, true))) {
        return PMIX_ERR_OUT_OF_RESOURCE;
    }

    /* add the new value - note that if the user is updating
     * a value, the ordering of the stored blobs will cause
     * an update to eventually occur. In other words, the
     * receiving process will first unpack the "old" data,
     * and then unpack the update and overwrite it */
    PMIX_RETAIN(kin);
    pmix_list_append(&proc_data->data, &kin->super);

    return PMIX_SUCCESS;
}

pmix_status_t pmix_hash_fetch(pmix_hash_table_t *table, int rank,
                              const char *key, pmix_value_t **kvs)
{
    pmix_status_t rc = PMIX_SUCCESS;
    pmix_proc_data_t *proc_data;
    pmix_kval_t *hv;
    uint64_t id;
    char *node;

    pmix_output_verbose(10, pmix_globals.debug_output,
                        "HASH:FETCH rank %d key %s",
                        rank, (NULL == key) ? "NULL" : key);

    id = (uint64_t)rank;

    if (PMIX_RANK_UNDEF == rank) {
        /* PMIX_RANK_UNDEF should return following statuses
         * PMIX_ERR_PROC_ENTRY_NOT_FOUND | PMIX_SUCCESS
         * special logic is basing on these statuses on a client and a server */
        rc = pmix_hash_table_get_first_key_uint64(table, &id,
                (void**)&proc_data, (void**)&node);
        if (PMIX_SUCCESS != rc) {
            pmix_output_verbose(10, pmix_globals.debug_output,
                                "HASH:FETCH proc data for rank %d not found",
                                rank);
            return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
        }
    } else {
        /* specified rank can return following statuses
         * PMIX_ERR_PROC_ENTRY_NOT_FOUND | PMIX_ERR_NOT_FOUND | PMIX_SUCCESS
         * special logic is basing on these statuses on a client and a server */
        id = (uint64_t)rank;    }

    while (PMIX_SUCCESS == rc) {
        proc_data = lookup_proc(table, id, false);
        if (NULL == proc_data) {
            pmix_output_verbose(10, pmix_globals.debug_output,
                                "HASH:FETCH proc data for rank %d not found",
                                rank);
            return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
        }

        /* if the key is NULL, then the user wants -all- data
         * put by the specified rank */
        if (NULL == key) {
            /* we will return the data as an array of pmix_info_t
             * in the kvs pmix_value_t */

        } else {
            /* find the value from within this proc_data object */
            hv = lookup_keyval(&proc_data->data, key);
            if (hv) {
                /* create the copy */
                if (PMIX_SUCCESS != (rc = pmix_bfrop.copy((void**)kvs, hv->value, PMIX_VALUE))) {
                    PMIX_ERROR_LOG(rc);
                    return rc;
                }
                break;
            } else if (PMIX_RANK_UNDEF != rank) {
                pmix_output_verbose(10, pmix_globals.debug_output,
                                    "HASH:FETCH data for key %s not found", key);
                return PMIX_ERR_NOT_FOUND;
            }
        }

        rc = pmix_hash_table_get_next_key_uint64(table, &id,
                (void**)&proc_data, node, (void**)&node);
        if (PMIX_SUCCESS != rc) {
            pmix_output_verbose(10, pmix_globals.debug_output,
                                "HASH:FETCH data for key %s not found", key);
            return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
        }
    }

    return rc;
}

pmix_status_t pmix_hash_fetch_by_key(pmix_hash_table_t *table, const char *key,
                                     int *rank, pmix_value_t **kvs, void **last)
{
    pmix_status_t rc = PMIX_SUCCESS;
    pmix_proc_data_t *proc_data;
    pmix_kval_t *hv;
    uint64_t id;
    char *node;
    static const char *key_r = NULL;

    if (key == NULL && (node = *last) == NULL) {
        return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
    }

    if (key == NULL && key_r == NULL) {
        return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
    }

    if (key) {
        rc = pmix_hash_table_get_first_key_uint64(table, &id,
                (void**)&proc_data, (void**)&node);
        key_r = key;
    } else {
        rc = pmix_hash_table_get_next_key_uint64(table, &id,
                (void**)&proc_data, node, (void**)&node);
    }

    pmix_output_verbose(10, pmix_globals.debug_output,
                        "HASH:FETCH BY KEY rank %d key %s",
                        (int)id, key_r);

    if (PMIX_SUCCESS != rc) {
        pmix_output_verbose(10, pmix_globals.debug_output,
                            "HASH:FETCH proc data for key %s not found",
                            key_r);
        return PMIX_ERR_PROC_ENTRY_NOT_FOUND;
    }

    /* find the value from within this proc_data object */
    hv = lookup_keyval(&proc_data->data, key_r);
    if (hv) {
        /* create the copy */
        if (PMIX_SUCCESS != (rc = pmix_bfrop.copy((void**)kvs, hv->value, PMIX_VALUE))) {
            PMIX_ERROR_LOG(rc);
            return rc;
        }
    } else {
        return PMIX_ERR_NOT_FOUND;
    }

    *rank = (int)id;
    *last = node;

    return PMIX_SUCCESS;
}

int pmix_hash_remove_data(pmix_hash_table_t *table,
                          int rank, const char *key)
{
    pmix_status_t rc = PMIX_SUCCESS;
    pmix_proc_data_t *proc_data;
    pmix_kval_t *kv;
    uint64_t id;
    char *node;

    id = (uint64_t)rank;

    /* if the rank is wildcard, we want to apply this to
     * all rank entries */
    if (PMIX_RANK_UNDEF == rank) {
        rc = pmix_hash_table_get_first_key_uint64(table, &id,
                (void**)&proc_data, (void**)&node);
        while (PMIX_SUCCESS == rc) {
            if (NULL != proc_data) {
                if (NULL == key) {
                    PMIX_RELEASE(proc_data);
                } else {
                    PMIX_LIST_FOREACH(kv, &proc_data->data, pmix_kval_t) {
                        if (0 == strcmp(key, kv->key)) {
                            pmix_list_remove_item(&proc_data->data, &kv->super);
                            PMIX_RELEASE(kv);
                            break;
                        }
                    }
                }
            }
            rc = pmix_hash_table_get_next_key_uint64(table, &id,
                    (void**)&proc_data, node, (void**)&node);
        }
    }

    /* lookup the specified proc */
    if (NULL == (proc_data = lookup_proc(table, id, false))) {
        /* no data for this proc */
        return PMIX_SUCCESS;
    }

    /* if key is NULL, remove all data for this proc */
    if (NULL == key) {
        while (NULL != (kv = (pmix_kval_t*)pmix_list_remove_first(&proc_data->data))) {
            PMIX_RELEASE(kv);
        }
        /* remove the proc_data object itself from the jtable */
        pmix_hash_table_remove_value_uint64(table, id);
        /* cleanup */
        PMIX_RELEASE(proc_data);
        return PMIX_SUCCESS;
    }

    /* remove this item */
    PMIX_LIST_FOREACH(kv, &proc_data->data, pmix_kval_t) {
        if (0 == strcmp(key, kv->key)) {
            pmix_list_remove_item(&proc_data->data, &kv->super);
            PMIX_RELEASE(kv);
            break;
        }
    }

    return PMIX_SUCCESS;
}

/**
 * Find data for a given key in a given pmix_list_t.
 */
static pmix_kval_t* lookup_keyval(pmix_list_t *data,
                                  const char *key)
{
    pmix_kval_t *kv;

    PMIX_LIST_FOREACH(kv, data, pmix_kval_t) {
        if (0 == strcmp(key, kv->key)) {
            return kv;
        }
    }
    return NULL;
}


/**
 * Find proc_data_t container associated with given
 * pmix_identifier_t.
 */
static pmix_proc_data_t* lookup_proc(pmix_hash_table_t *jtable,
                                     uint64_t id, bool create)
{
    pmix_proc_data_t *proc_data = NULL;

    pmix_hash_table_get_value_uint64(jtable, id, (void**)&proc_data);
    if (NULL == proc_data && create) {
        /* The proc clearly exists, so create a data structure for it */
        proc_data = PMIX_NEW(pmix_proc_data_t);
        if (NULL == proc_data) {
            pmix_output(0, "pmix:client:hash:lookup_pmix_proc: unable to allocate proc_data_t\n");
            return NULL;
        }
        pmix_hash_table_set_value_uint64(jtable, id, proc_data);
    }

    return proc_data;
}
