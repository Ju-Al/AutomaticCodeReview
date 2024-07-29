      return timestamp;
    }

    public String getTimestampISO() {
      return tsFormatter.format(timestamp);
    }

    public boolean isCompressed() {
      return compressed;
    }

    public void writeHeader(final DataOutputStream dos) throws IOException {
      dos.writeUTF(encodingVer);
      dos.writeInt(dataVersion);
      dos.writeUTF(tsFormatter.format(timestamp));
      dos.writeBoolean(compressed);
    }

    @Override
    public String toString() {
      return "Header{" + "encodingVer='" + encodingVer + '\'' + ", dataVersion=" + dataVersion
          + ", timestamp=" + timestamp + ", compressed=" + compressed + '}';
    }
  }

}
