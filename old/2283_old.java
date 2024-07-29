 */
public class CompactionInfo {

  // Variable names become JSON keys
  public String server;

  public long count;
  public Long oldest;

  public CompactionInfo() {}

  /**
   * Stores new compaction information
   *
   * @param tserverInfo
   *          status of the tserver
   * @param count
   *          number of compactions
   * @param oldest
   *          time of oldest compaction
   */
  public CompactionInfo(TabletServerStatus tserverInfo, long count, Long oldest) {
    this.server = tserverInfo.getName();
    this.count = count;
    this.oldest = oldest;
  }
}
