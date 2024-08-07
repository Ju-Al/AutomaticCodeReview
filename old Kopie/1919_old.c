/*
 *
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *
 *  http://aws.amazon.com/apache2.0
 *
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

#include <sys/param.h>

#include "utils/s2n_mem.h"
#include "utils/s2n_random.h"
#include "utils/s2n_safety.h"
#include "crypto/s2n_drbg.h"
#include "tests/testlib/s2n_testlib.h"

#if S2N_LIBCRYPTO_SUPPORTS_CUSTOM_RAND

static uint8_t seed_buffer[S2N_DRBG_MAX_SEED_SIZE] = { 0 };

S2N_RESULT s2n_fixed_entropy_generator(struct s2n_blob *blob)
{
    ENSURE_LTE(blob->size, S2N_DRBG_MAX_SEED_SIZE);
    blob->data = seed_buffer;
    return S2N_RESULT_OK;
}

static struct s2n_drbg fixed_drbg = { .entropy_generator = &s2n_fixed_entropy_generator };

int s2n_unsafe_drbg_reseed(uint8_t *seed, uint8_t seed_size)
{
    memset(seed_buffer, 0, S2N_DRBG_MAX_SEED_SIZE);
    memcpy_check(seed_buffer, seed, MIN(seed_size, S2N_DRBG_MAX_SEED_SIZE));

    uint8_t data[48] = { 0 };
    struct s2n_blob ps;
    GUARD(s2n_blob_init(&ps, data, sizeof(data)));

    GUARD(s2n_drbg_instantiate(&fixed_drbg, &ps, S2N_DANGEROUS_AES_256_CTR_NO_DF_NO_PR));
    GUARD_AS_POSIX(s2n_set_private_drbg_for_test(fixed_drbg));

    return S2N_SUCCESS;
}

#else

int s2n_unsafe_drbg_reseed(uint8_t *seed, uint8_t seed_size)
{
    S2N_ERROR(S2N_ERR_UNIMPLEMENTED);
}

#endif
