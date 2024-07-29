    return nullptr;
  }
  // Skip table slice data already processed.
  auto bytes_read = static_cast<size_t>(buf.in_avail());
  VAST_ASSERT(chunk->size() > bytes_read);
  auto header_size = chunk->size() - bytes_read;
  if (auto err = result.unshared().load(chunk->slice(header_size))) {
    VAST_ERROR_ANON(__func__, "failed to load table slice from chunk");
    return nullptr;
  }
  return result;
}

} // namespace vast
