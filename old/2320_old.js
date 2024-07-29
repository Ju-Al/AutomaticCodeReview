            onSave();
        }
    }, [onSave]);

    const showEditModal = useCallback(() => {
        toggleDrawer('edit.payment');
        setIsEditModalHidden(false);
    }, [setIsEditModalHidden, toggleDrawer]);

    const hideEditModal = useCallback(() => {
        setIsEditModalHidden(true);
        closeDrawer('edit.payment');
    }, [setIsEditModalHidden, closeDrawer]);

    return {
        doneEditing: !!paymentNonce,
        handleReviewOrder,
        shouldRequestPaymentNonce,
        selectedPaymentMethod,
        paymentNonce,
        isEditModalHidden,
        showEditModal,
        hideEditModal
    };
};
