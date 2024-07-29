    }

    private void addRecipientFromContactUri(final RecipientType recipientType, final Uri uri) {
        new RecipientLoader(context, cryptoProvider, uri) {
            @Override
            public void deliverResult(List<Recipient> result) {
                // TODO handle multiple available mail addresses for a contact?
                if (result.isEmpty()) {
                    recipientView.showErrorContactNoAddress();
                    return;
                }
                Recipient recipient = result.get(0);
                recipientView.addRecipients(recipientType, recipient);
                stopLoading();
            }
        }.startLoading();
    }

    public void onToFocused() {
        lastFocusedType = RecipientType.TO;
    }

    public void onCcFocused() {
        lastFocusedType = RecipientType.CC;
    }

    public void onBccFocused() {
        lastFocusedType = RecipientType.BCC;
    }

    public void onMenuAddFromContacts() {
        int requestCode = recipientTypeToRequestCode(lastFocusedType);
        recipientView.showContactPicker(requestCode);
    }

    public void onActivityResult(int requestCode, Intent data) {
        if (requestCode != CONTACT_PICKER_TO && requestCode != CONTACT_PICKER_CC && requestCode != CONTACT_PICKER_BCC) {
            return;
        }

        RecipientType recipientType = recipientTypeFromRequestCode(requestCode);
        addRecipientFromContactUri(recipientType, data.getData());
    }

    private static final int CONTACT_PICKER_TO = 1;
    private static final int CONTACT_PICKER_CC = 2;
    private static final int CONTACT_PICKER_BCC = 3;

    private static int recipientTypeToRequestCode(RecipientType type) {
        switch (type) {
            case TO:
            default:
                return CONTACT_PICKER_TO;
            case CC:
                return CONTACT_PICKER_CC;
            case BCC:
                return CONTACT_PICKER_BCC;
        }

    }

    private static RecipientType recipientTypeFromRequestCode(int type) {
        switch (type) {
            case CONTACT_PICKER_TO:
            default:
                return RecipientType.TO;
            case CONTACT_PICKER_CC:
                return RecipientType.CC;
            case CONTACT_PICKER_BCC:
                return RecipientType.BCC;
        }

    }

    public void onNonRecipientFieldFocused() {
        hideEmptyExtendedRecipientFields();
    }

    public void onClickCryptoStatus() {
        recipientView.showCryptoDialog(currentCryptoMode);
    }

    public void onCryptoModeChanged(CryptoMode type) {
        currentCryptoMode = type;
        updateCryptoDisplayStatus();
    }

    enum CryptoMode {
        OPPORTUNISTIC, DISABLE, SIGN_ONLY
    }

}
