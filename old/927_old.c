//////////////////////////////////////

// Return 1 if set
_INLINE_ int is_bit_set(IN const uint64_t *a,
                        IN const uint32_t pos)
{
    
    const uint32_t qw_pos = pos >> 6;
    const uint32_t bit_pos = pos & 0x3f;

    return ((a[qw_pos] & BIT(bit_pos)) != 0);
}

_INLINE_ void set_bit(IN uint64_t *a,
                      IN const uint32_t pos)
{
    
    const uint32_t qw_pos = pos >> 6;
    const uint32_t bit_pos = pos & 0x3f;

    a[qw_pos] |= BIT(bit_pos);
}

// Assumption fake_weight > waight
status_t generate_sparse_fake_rep(OUT uint64_t *a,
                                  OUT idx_t wlist[],
                                  IN  const uint32_t padded_len,
                                  IN OUT aes_ctr_prf_state_t *prf_state)
{
    const uint32_t len = R_BITS;

#ifndef VALIDATE_CONSTANT_TIME
    return generate_sparse_rep(a, wlist, DV, len, padded_len, prf_state);
#else
    // This part of code to be able to compare constant and 
    // non constant time implementations.

    status_t res = SUCCESS;
    uint64_t ctr = 0;
    
    uint32_t real_wlist[DV];
    idx_t inordered_wlist[FAKE_DV];

    GUARD(generate_sparse_rep(a, inordered_wlist, FAKE_DV, 
                              len, padded_len, prf_state), res, EXIT);

    // Initialize to zero
    memset(a, 0, (len + 7) >> 3);

    // Allocate DV real positions
    do
    {
        GUARD(get_rand_mod_len(&real_wlist[ctr], FAKE_DV, prf_state), res, EXIT);

        wlist[ctr].val = inordered_wlist[real_wlist[ctr]].val;
        set_bit(a, wlist[ctr].val);

        ctr += is_new2(real_wlist, ctr);
    } while(ctr < DV);

EXIT:
    return res;
#endif
}

// Assumption fake_weight > waight
status_t generate_sparse_rep(OUT uint64_t *a,
                             OUT idx_t wlist[],
                             IN  const uint32_t weight,
                             IN  const uint32_t len,
                             IN  const uint32_t padded_len,
                             IN OUT aes_ctr_prf_state_t *prf_state)
{
    BIKE_UNUSED(padded_len);

    status_t res = SUCCESS;
    uint64_t ctr = 0;

    // Initialize to zero
    memset(a, 0, (len + 7) >> 3);

    do
    {
         GUARD(get_rand_mod_len(&wlist[ctr].val, len, prf_state), res, EXIT);

         if (!is_bit_set(a, wlist[ctr].val))
         {
             set_bit(a, wlist[ctr].val);
             ctr++;
         }
    } while(ctr < weight);

EXIT:
    return res;
}

#endif
