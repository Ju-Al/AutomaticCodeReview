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
