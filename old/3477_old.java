        super(context);
    }

    @Override
    protected Uri onRestoreRingtone() {
        return currentRingtone;
    }

    public void setCurrentRingtone (Uri uri) {
        currentRingtone = uri;
    }
}
