    this.downloadListener = listener;
  }

  @Override
  public void onStart() {
    this.controlToggle.display(this.pauseButton);
  }

  @Override
  public void onStop() {
    this.controlToggle.display(this.playButton);
  }

  @Override
  public void onProgress(double progress, long millis) {
    int seekProgress = (int)Math.floor(progress * this.seekBar.getMax());

    if (seekProgress > seekBar.getProgress() || backwardsCounter > 3) {
      backwardsCounter = 0;
      this.seekBar.setProgress(seekProgress);
      this.timestamp.setText(String.format("%02d:%02d",
                                           TimeUnit.MILLISECONDS.toMinutes(millis),
                                           TimeUnit.MILLISECONDS.toSeconds(millis)));
    } else {
      backwardsCounter++;
    }
  }

  public void setTint(int tint) {
    this.playButton.setColorFilter(tint, PorterDuff.Mode.SRC_IN);
    this.pauseButton.setColorFilter(tint, PorterDuff.Mode.SRC_IN);
    this.downloadButton.setColorFilter(tint, PorterDuff.Mode.SRC_IN);
    this.downloadProgress.setBarColor(tint);

    this.timestamp.setTextColor(tint);
    this.seekBar.getProgressDrawable().setColorFilter(tint, PorterDuff.Mode.SRC_IN);

    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.JELLY_BEAN) {
      this.seekBar.getThumb().setColorFilter(tint, PorterDuff.Mode.SRC_IN);
    }
  }

  private double getProgress() {
    if (this.seekBar.getProgress() <= 0 || this.seekBar.getMax() <= 0) {
      return 0;
    } else {
      return (double)this.seekBar.getProgress() / (double)this.seekBar.getMax();
    }
  }

  private class PlayClickedListener implements View.OnClickListener {
    @Override
    public void onClick(View v) {
      try {
        Log.w(TAG, "playbutton onClick");
        if (audioSlidePlayer != null) {
          controlToggle.display(pauseButton);
          audioSlidePlayer.play(getProgress());
        }
      } catch (IOException e) {
        Log.w(TAG, e);
      }
    }
  }

  private class PauseClickedListener implements View.OnClickListener {
    @Override
    public void onClick(View v) {
      Log.w(TAG, "pausebutton onClick");
      if (audioSlidePlayer != null) {
        controlToggle.display(playButton);
        audioSlidePlayer.stop();
      }
    }
  }

  private class DownloadClickedListener implements View.OnClickListener {
    private final @NonNull AudioSlide slide;

    private DownloadClickedListener(@NonNull AudioSlide slide) {
      this.slide = slide;
    }

    @Override
    public void onClick(View v) {
      if (downloadListener != null) downloadListener.onClick(v, slide);
    }
  }

  private class SeekBarModifiedListener implements SeekBar.OnSeekBarChangeListener {
    @Override
    public void onProgressChanged(SeekBar seekBar, int progress, boolean fromUser) {}

    @Override
    public synchronized void onStartTrackingTouch(SeekBar seekBar) {
      if (audioSlidePlayer != null && pauseButton.getVisibility() == View.VISIBLE) {
        audioSlidePlayer.stop();
      }
    }

    @Override
    public synchronized void onStopTrackingTouch(SeekBar seekBar) {
      try {
        if (audioSlidePlayer != null && pauseButton.getVisibility() == View.VISIBLE) {
          audioSlidePlayer.play(getProgress());
        }
      } catch (IOException e) {
        Log.w(TAG, e);
      }
    }
  }

  @SuppressWarnings("unused")
  public void onEventAsync(final PartProgressEvent event) {
    if (audioSlidePlayer != null && event.attachment.equals(this.audioSlidePlayer.getAudioSlide().asAttachment())) {
      Util.runOnMain(new Runnable() {
        @Override
        public void run() {
          downloadProgress.setInstantProgress(((float) event.progress) / event.total);
        }
      });
    }
  }

}
