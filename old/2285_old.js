
    const handleSubmit = useCallback(
        async formValues => {
            const { country, email, ...address } = formValues;
            await setShippingInformation({
                variables: {
                    cartId,
                    email,
                    address: {
                        ...address,
                        country_code: country
                    }
                }
            });

            if (afterSubmit) {
                afterSubmit();
            }
        },
        [afterSubmit, cartId, setShippingInformation]
    );

    return {
        handleSubmit,
        initialValues,
        isSaving: called && loading,
        isUpdate
    };
};
