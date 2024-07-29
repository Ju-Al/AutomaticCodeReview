                // Found contact, so trigger QuickContact
                QuickContact.showQuickContact(
                        getContext(), ContactBadge.this, lookupUri, QuickContact.MODE_LARGE, null);
            } else if (createUri != null) {
                // Prompt user to add this person to contacts
                final Intent intent = new Intent(Intents.SHOW_OR_CREATE_CONTACT, createUri);
                extras.remove(EXTRA_URI_CONTENT);
                intent.putExtras(extras);
                getContext().startActivity(intent);
            }
        }
    }
}

