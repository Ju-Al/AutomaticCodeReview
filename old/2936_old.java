                                    if (message.isSet(Flag.DELETED)) {
                                        Timber.v("Newly downloaded message %s:%s:%s was marked deleted on server, " +
                                                "skipping", account, folder, message.getUid());
                                    } else {
                                        Timber.d("Newly downloaded message %s is older than %s, skipping",
                                                message.getUid(), earliestDate);
                                    }
                                }
                                progress.incrementAndGet();
                                for (MessagingListener l : controller.getListeners()) {
                                    //TODO: This might be the source of poll count errors in the UI. Is todo always the
                                    // same as ofTotal
                                    l.synchronizeMailboxProgress(account, folder, progress.get(), todo);
                                }
                                return;
                            }

                            if (account.getMaximumAutoDownloadMessageSize() > 0 &&
                                    message.getSize() > account.getMaximumAutoDownloadMessageSize()) {
                                largeMessages.add(message);
                            } else {
                                smallMessages.add(message);
                            }
                        } catch (Exception e) {
                            Timber.e(e, "Error while storing downloaded message.");
                        }
                    }

                    @Override
                    public void messageStarted(String uid, int number, int ofTotal) {
                    }

                    @Override
                    public void messagesFinished(int total) {
                        // FIXME this method is almost never invoked by various Stores! Don't rely on it unless fixed!!
                    }

                });
    }

    private boolean shouldImportMessage(final Account account, final Message message,
            final Date earliestDate) {

        if (account.isSearchByDateCapable() && message.olderThan(earliestDate)) {
            Timber.d("Message %s is older than %s, hence not saving", message.getUid(), earliestDate);
            return false;
        }
        return true;
    }

    private <T extends Message> void downloadSmallMessages(final Account account, final Folder<T> remoteFolder,
            final LocalFolder localFolder,
            List<T> smallMessages,
            final AtomicInteger progress,
            final int unreadBeforeStart,
            final AtomicInteger newMessages,
            final int todo,
            FetchProfile fp) throws MessagingException {
        final String folder = remoteFolder.getName();

        final Date earliestDate = account.getEarliestPollDate();

        Timber.d("SYNC: Fetching %d small messages for folder %s", smallMessages.size(), folder);

        remoteFolder.fetch(smallMessages,
                fp, new MessageRetrievalListener<T>() {
                    @Override
                    public void messageFinished(final T message, int number, int ofTotal) {
                        try {

                            if (!shouldImportMessage(account, message, earliestDate)) {
                                progress.incrementAndGet();

                                return;
                            }

                            // Store the updated message locally
                            final LocalMessage localMessage = localFolder.storeSmallMessage(message, new Runnable() {
                                @Override
                                public void run() {
                                    progress.incrementAndGet();
                                }
                            });

                            // Increment the number of "new messages" if the newly downloaded message is
                            // not marked as read.
                            if (!localMessage.isSet(Flag.SEEN)) {
                                newMessages.incrementAndGet();
                            }

                            Timber.v("About to notify listeners that we got a new small message %s:%s:%s",
                                    account, folder, message.getUid());

                            // Update the listener with what we've found
                            for (MessagingListener l : controller.getListeners()) {
                                l.synchronizeMailboxProgress(account, folder, progress.get(), todo);
                                if (!localMessage.isSet(Flag.SEEN)) {
                                    l.synchronizeMailboxNewMessage(account, folder, localMessage);
                                }
                            }
                            // Send a notification of this message

                            if (syncHelper.shouldNotifyForMessage(account, localFolder, message, contacts)) {
                                // Notify with the localMessage so that we don't have to recalculate the content preview.
                                notificationController.addNewMailNotification(account, localMessage, unreadBeforeStart);
                            }

                        } catch (MessagingException me) {
                            Timber.e(me, "SYNC: fetch small messages");
                        }
                    }

                    @Override
                    public void messageStarted(String uid, int number, int ofTotal) {
                    }

                    @Override
                    public void messagesFinished(int total) {
                    }
                });

        Timber.d("SYNC: Done fetching small messages for folder %s", folder);
    }

    private <T extends Message> void downloadLargeMessages(final Account account, final Folder<T> remoteFolder,
            final LocalFolder localFolder,
            List<T> largeMessages,
            final AtomicInteger progress,
            final int unreadBeforeStart,
            final AtomicInteger newMessages,
            final int todo,
            FetchProfile fp) throws MessagingException {
        final String folder = remoteFolder.getName();
        final Date earliestDate = account.getEarliestPollDate();

        Timber.d("SYNC: Fetching large messages for folder %s", folder);

        remoteFolder.fetch(largeMessages, fp, null);
        for (T message : largeMessages) {

            if (!shouldImportMessage(account, message, earliestDate)) {
                progress.incrementAndGet();
                continue;
            }

            if (message.getBody() == null) {
                downloadSaneBody(account, remoteFolder, localFolder, message);
            } else {
                downloadPartial(remoteFolder, localFolder, message);
            }

            Timber.v("About to notify listeners that we got a new large message %s:%s:%s",
                    account, folder, message.getUid());

            // Update the listener with what we've found
            progress.incrementAndGet();
            // TODO do we need to re-fetch this here?
            LocalMessage localMessage = localFolder.getMessage(message.getUid());
            // Increment the number of "new messages" if the newly downloaded message is
            // not marked as read.
            if (!localMessage.isSet(Flag.SEEN)) {
                newMessages.incrementAndGet();
            }
            for (MessagingListener l : controller.getListeners()) {
                l.synchronizeMailboxProgress(account, folder, progress.get(), todo);
                if (!localMessage.isSet(Flag.SEEN)) {
                    l.synchronizeMailboxNewMessage(account, folder, localMessage);
                }
            }
            // Send a notification of this message
            if (syncHelper.shouldNotifyForMessage(account, localFolder, message, contacts)) {
                // Notify with the localMessage so that we don't have to recalculate the content preview.
                notificationController.addNewMailNotification(account, localMessage, unreadBeforeStart);
            }
        }

        Timber.d("SYNC: Done fetching large messages for folder %s", folder);
    }

    private void downloadPartial(Folder remoteFolder, LocalFolder localFolder, Message message)
            throws MessagingException {
        /*
         * We have a structure to deal with, from which
         * we can pull down the parts we want to actually store.
         * Build a list of parts we are interested in. Text parts will be downloaded
         * right now, attachments will be left for later.
         */

        Set<Part> viewables = MessageExtractor.collectTextParts(message);

        /*
         * Now download the parts we're interested in storing.
         */
        BodyFactory bodyFactory = new DefaultBodyFactory();
        for (Part part : viewables) {
            remoteFolder.fetchPart(message, part, null, bodyFactory);
        }
        // Store the updated message locally
        localFolder.appendMessages(Collections.singletonList(message));

        Message localMessage = localFolder.getMessage(message.getUid());

        localMessage.setFlag(Flag.X_DOWNLOADED_PARTIAL, true);
    }

    private void downloadSaneBody(Account account, Folder remoteFolder, LocalFolder localFolder, Message message)
            throws MessagingException {
        /*
         * The provider was unable to get the structure of the message, so
         * we'll download a reasonable portion of the messge and mark it as
         * incomplete so the entire thing can be downloaded later if the user
         * wishes to download it.
         */
        FetchProfile fp = new FetchProfile();
        fp.add(FetchProfile.Item.BODY_SANE);
                /*
                 *  TODO a good optimization here would be to make sure that all Stores set
                 *  the proper size after this fetch and compare the before and after size. If
                 *  they equal we can mark this SYNCHRONIZED instead of PARTIALLY_SYNCHRONIZED
                 */

        remoteFolder.fetch(Collections.singletonList(message), fp, null);

        // Store the updated message locally
        localFolder.appendMessages(Collections.singletonList(message));

        Message localMessage = localFolder.getMessage(message.getUid());


        // Certain (POP3) servers give you the whole message even when you ask for only the first x Kb
        if (!message.isSet(Flag.X_DOWNLOADED_FULL)) {
                    /*
                     * Mark the message as fully downloaded if the message size is smaller than
                     * the account's autodownload size limit, otherwise mark as only a partial
                     * download.  This will prevent the system from downloading the same message
                     * twice.
                     *
                     * If there is no limit on autodownload size, that's the same as the message
                     * being smaller than the max size
                     */
            if (account.getMaximumAutoDownloadMessageSize() == 0
                    || message.getSize() < account.getMaximumAutoDownloadMessageSize()) {
                localMessage.setFlag(Flag.X_DOWNLOADED_FULL, true);
            } else {
                // Set a flag indicating that the message has been partially downloaded and
                // is ready for view.
                localMessage.setFlag(Flag.X_DOWNLOADED_PARTIAL, true);
            }
        }

    }
}