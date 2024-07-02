#include "gamesmodel.h"

    QString ret;
    if (secs < SECS_PER_MIN * 2) // for first min we display "New"
        ret = tr("New");
    else if (secs < SECS_PER_MIN * 10) // from 2 - 10 mins we show the mins
        ret = QString("%1 min").arg(QString::number(secs / SECS_PER_MIN));
    else if (secs < SECS_PER_MIN * 60) { // from 10 mins to 1h we aggregate every 10 mins
        int unitOfTen = secs / SECS_PER_TEN_MIN;
        QString str = "%1%2";
        ret = str.arg(QString::number(unitOfTen), "0+ min");
    } else { // from 1 hr onward we show hrs
        int hours = secs / SECS_PER_HOUR;
        if (secs % SECS_PER_HOUR >= SECS_PER_MIN * 30) // if the room is open for 1hr 30 mins, we round to 2hrs
            ++hours;
        ret = QString("%1+ h").arg(QString::number(hours));
#include "pb/serverinfo_game.pb.h"
#include "pixmapgenerator.h"
#include "settingscache.h"
#include "tab_userlists.h"
#include "userlist.h"

#include <QDateTime>
#include <QDebug>
#include <QIcon>
#include <QStringList>
#include <QTime>

enum GameListColumn
{
    ROOM,
    CREATED,
    DESCRIPTION,
    CREATOR,
    GAME_TYPE,
    RESTRICTIONS,
    PLAYERS,
    SPECTATORS
};

const QString GamesModel::getGameCreatedString(const int secs) const
{
    static const QTime zeroTime{0, 0};
    static const int wrapSeconds = zeroTime.secsTo(zeroTime.addSecs(-1));
    static const int halfHourSecs = zeroTime.secsTo(QTime(1, 0)) / 2;
    static const int halfMinSecs = zeroTime.secsTo(QTime(0, 1)) / 2;

    if (secs > wrapSeconds) { // QTime wraps after a day
        return tr("Days");
    }

    QTime total = zeroTime.addSecs(secs);
    if (total.hour()) {
        total = total.addSecs(halfHourSecs); // round up
        return QString::number(total.hour()) + QLatin1String("+ h");
    } else if (total.minute() < 2) { // games are new during their first minute
        return tr("New");
    } else {
        total = total.addSecs(halfMinSecs); // round up
        int tenMinutes = total.minute() / 10;
        if (tenMinutes) { // aggregate multiples of ten
            return QString::number(tenMinutes) + QLatin1String("0+ m");
        } else {
            return QString::number(total.minute()) + QLatin1String("+ m");
        }
    }
}

GamesModel::GamesModel(const QMap<int, QString> &_rooms, const QMap<int, GameTypeMap> &_gameTypes, QObject *parent)
    : QAbstractTableModel(parent), rooms(_rooms), gameTypes(_gameTypes)
{
}

QVariant GamesModel::data(const QModelIndex &index, int role) const
{
    if (!index.isValid())
        return QVariant();
    if (role == Qt::UserRole)
        return index.row();
    if (role != Qt::DisplayRole && role != SORT_ROLE && role != Qt::DecorationRole && role != Qt::TextAlignmentRole)
        return QVariant();
    if ((index.row() >= gameList.size()) || (index.column() >= columnCount()))
        return QVariant();

    const ServerInfo_Game &g = gameList[index.row()];
    switch (index.column()) {
        case ROOM:
            return rooms.value(g.room_id());
        case CREATED: {
            QDateTime then;
            then.setTime_t(g.start_time());
            int secs = then.secsTo(QDateTime::currentDateTime());

            switch (role) {
                case Qt::DisplayRole:
                    return getGameCreatedString(secs);
                case SORT_ROLE:
                    return QVariant(secs);
                case Qt::TextAlignmentRole:
                    return Qt::AlignCenter;
                default:
                    return QVariant();
            }
        }
        case DESCRIPTION:
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole:
                    return QString::fromStdString(g.description());
                case Qt::TextAlignmentRole:
                    return Qt::AlignLeft;
                default:
                    return QVariant();
            }
        case CREATOR: {
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole:
                    return QString::fromStdString(g.creator_info().name());
                case Qt::DecorationRole: {
                    QPixmap avatarPixmap = UserLevelPixmapGenerator::generatePixmap(
                        13, (UserLevelFlags)g.creator_info().user_level(), false,
                        QString::fromStdString(g.creator_info().privlevel()));
                    return QIcon(avatarPixmap);
                }
                default:
                    return QVariant();
            }
        }
        case GAME_TYPE:
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole: {
                    QStringList result;
                    GameTypeMap gameTypeMap = gameTypes.value(g.room_id());
                    for (int i = g.game_types_size() - 1; i >= 0; --i)
                        result.append(gameTypeMap.value(g.game_types(i)));
                    return result.join(", ");
                }
                case Qt::TextAlignmentRole:
                    return Qt::AlignLeft;
                default:
                    return QVariant();
            }
        case RESTRICTIONS:
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole: {
                    QStringList result;
                    if (g.with_password())
                        result.append(tr("password"));
                    if (g.only_buddies())
                        result.append(tr("buddies only"));
                    if (g.only_registered())
                        result.append(tr("reg. users only"));
                    return result.join(", ");
                }
                case Qt::DecorationRole: {
                    return g.with_password() ? QIcon(LockPixmapGenerator::generatePixmap(13)) : QVariant();
                    case Qt::TextAlignmentRole:
                        return Qt::AlignLeft;
                    default:
                        return QVariant();
                }
            }
        case PLAYERS:
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole:
                    return QString("%1/%2").arg(g.player_count()).arg(g.max_players());
                case Qt::TextAlignmentRole:
                    return Qt::AlignCenter;
                default:
                    return QVariant();
            }

        case SPECTATORS:
            switch (role) {
                case SORT_ROLE:
                case Qt::DisplayRole: {
                    if (g.spectators_allowed()) {
                        QString result;
                        result.append(QString::number(g.spectators_count()));

                        if (g.spectators_can_chat() && g.spectators_omniscient()) {
                            result.append(" (")
                                .append(tr("can chat"))
                                .append(" & ")
                                .append(tr("see hands"))
                                .append(")");
                        } else if (g.spectators_can_chat()) {
                            result.append(" (").append(tr("can chat")).append(")");
                        } else if (g.spectators_omniscient()) {
                            result.append(" (").append(tr("can see hands")).append(")");
                        }

                        return result;
                    }
                    return QVariant(tr("not allowed"));
                }
                case Qt::TextAlignmentRole:
                    return Qt::AlignLeft;
                default:
                    return QVariant();
            }
        default:
            return QVariant();
    }
}

QVariant GamesModel::headerData(int section, Qt::Orientation /*orientation*/, int role) const
{
    if ((role != Qt::DisplayRole) && (role != Qt::TextAlignmentRole))
        return QVariant();
    switch (section) {
        case ROOM:
            return tr("Room");
        case CREATED: {
            switch (role) {
                case Qt::DisplayRole:
                    return tr("Age");
                case Qt::TextAlignmentRole:
                    return Qt::AlignCenter;
                default:
                    return QVariant();
            }
        }
        case DESCRIPTION:
            return tr("Description");
        case CREATOR:
            return tr("Creator");
        case GAME_TYPE:
            return tr("Type");
        case RESTRICTIONS:
            return tr("Restrictions");
        case PLAYERS: {
            switch (role) {
                case Qt::DisplayRole:
                    return tr("Players");
                case Qt::TextAlignmentRole:
                    return Qt::AlignCenter;
                default:
                    return QVariant();
            }
        }
        case SPECTATORS:
            return tr("Spectators");
        default:
            return QVariant();
    }
}

const ServerInfo_Game &GamesModel::getGame(int row)
{
    Q_ASSERT(row < gameList.size());
    return gameList[row];
}

void GamesModel::updateGameList(const ServerInfo_Game &game)
{
    for (int i = 0; i < gameList.size(); i++) {
        if (gameList[i].game_id() == game.game_id()) {
            if (game.closed()) {
                beginRemoveRows(QModelIndex(), i, i);
                gameList.removeAt(i);
                endRemoveRows();
            } else {
                gameList[i].MergeFrom(game);
                emit dataChanged(index(i, 0), index(i, NUM_COLS - 1));
            }
            return;
        }
    }
    if (game.player_count() <= 0)
        return;
    beginInsertRows(QModelIndex(), gameList.size(), gameList.size());
    gameList.append(game);
    endInsertRows();
}

GamesProxyModel::GamesProxyModel(QObject *parent, const TabSupervisor *_tabSupervisor)
    : QSortFilterProxyModel(parent), ownUserIsRegistered(_tabSupervisor->isOwnUserRegistered()),
      tabSupervisor(_tabSupervisor)
{
    resetFilterParameters();
    setSortRole(GamesModel::SORT_ROLE);
    setDynamicSortFilter(true);
}

void GamesProxyModel::setShowBuddiesOnlyGames(bool _showBuddiesOnlyGames)
{
    showBuddiesOnlyGames = _showBuddiesOnlyGames;
    invalidateFilter();
}

void GamesProxyModel::setHideIgnoredUserGames(bool _hideIgnoredUserGames)
{
    hideIgnoredUserGames = _hideIgnoredUserGames;
    invalidateFilter();
}

void GamesProxyModel::setUnavailableGamesVisible(bool _unavailableGamesVisible)
{
    unavailableGamesVisible = _unavailableGamesVisible;
    invalidateFilter();
}

void GamesProxyModel::setShowPasswordProtectedGames(bool _showPasswordProtectedGames)
{
    showPasswordProtectedGames = _showPasswordProtectedGames;
    invalidateFilter();
}

void GamesProxyModel::setGameNameFilter(const QString &_gameNameFilter)
{
    gameNameFilter = _gameNameFilter;
    invalidateFilter();
}

void GamesProxyModel::setCreatorNameFilter(const QString &_creatorNameFilter)
{
    creatorNameFilter = _creatorNameFilter;
    invalidateFilter();
}

void GamesProxyModel::setGameTypeFilter(const QSet<int> &_gameTypeFilter)
{
    gameTypeFilter = _gameTypeFilter;
    invalidateFilter();
}

void GamesProxyModel::setMaxPlayersFilter(int _maxPlayersFilterMin, int _maxPlayersFilterMax)
{
    maxPlayersFilterMin = _maxPlayersFilterMin;
    maxPlayersFilterMax = _maxPlayersFilterMax;
    invalidateFilter();
}

int GamesProxyModel::getNumFilteredGames() const
{
    GamesModel *model = qobject_cast<GamesModel *>(sourceModel());
    if (!model)
        return 0;

    int numFilteredGames = 0;
    for (int row = 0; row < model->rowCount(); ++row) {
        if (!filterAcceptsRow(row)) {
            ++numFilteredGames;
        }
    }
    return numFilteredGames;
}

void GamesProxyModel::resetFilterParameters()
{
    unavailableGamesVisible = DEFAULT_UNAVAILABLE_GAMES_VISIBLE;
    showPasswordProtectedGames = DEFAULT_SHOW_PASSWORD_PROTECTED_GAMES;
    showBuddiesOnlyGames = DEFAULT_SHOW_BUDDIES_ONLY_GAMES;
    hideIgnoredUserGames = DEFAULT_HIDE_IGNORED_USER_GAMES;
    gameNameFilter = QString();
    creatorNameFilter = QString();
    gameTypeFilter.clear();
    maxPlayersFilterMin = DEFAULT_MAX_PLAYERS_MIN;
    maxPlayersFilterMax = DEFAULT_MAX_PLAYERS_MAX;

    invalidateFilter();
}

bool GamesProxyModel::areFilterParametersSetToDefaults() const
{
    return unavailableGamesVisible == DEFAULT_UNAVAILABLE_GAMES_VISIBLE &&
           showPasswordProtectedGames == DEFAULT_SHOW_PASSWORD_PROTECTED_GAMES &&
           showBuddiesOnlyGames == DEFAULT_SHOW_BUDDIES_ONLY_GAMES &&
           hideIgnoredUserGames == DEFAULT_HIDE_IGNORED_USER_GAMES && gameNameFilter.isEmpty() &&
           creatorNameFilter.isEmpty() && gameTypeFilter.isEmpty() && maxPlayersFilterMin == DEFAULT_MAX_PLAYERS_MIN &&
           maxPlayersFilterMax == DEFAULT_MAX_PLAYERS_MAX;
}

void GamesProxyModel::loadFilterParameters(const QMap<int, QString> &allGameTypes)
{
    GameFiltersSettings &gameFilters = SettingsCache::instance().gameFilters();
    unavailableGamesVisible = gameFilters.isUnavailableGamesVisible();
    showPasswordProtectedGames = gameFilters.isShowPasswordProtectedGames();
    hideIgnoredUserGames = gameFilters.isHideIgnoredUserGames();
    gameNameFilter = gameFilters.getGameNameFilter();
    maxPlayersFilterMin = gameFilters.getMinPlayers();
    maxPlayersFilterMax = gameFilters.getMaxPlayers();

    QMapIterator<int, QString> gameTypesIterator(allGameTypes);
    while (gameTypesIterator.hasNext()) {
        gameTypesIterator.next();
        if (gameFilters.isGameTypeEnabled(gameTypesIterator.value())) {
            gameTypeFilter.insert(gameTypesIterator.key());
        }
    }

    invalidateFilter();
}

void GamesProxyModel::saveFilterParameters(const QMap<int, QString> &allGameTypes)
{
    GameFiltersSettings &gameFilters = SettingsCache::instance().gameFilters();
    gameFilters.setShowBuddiesOnlyGames(showBuddiesOnlyGames);
    gameFilters.setUnavailableGamesVisible(unavailableGamesVisible);
    gameFilters.setShowPasswordProtectedGames(showPasswordProtectedGames);
    gameFilters.setHideIgnoredUserGames(hideIgnoredUserGames);
    gameFilters.setGameNameFilter(gameNameFilter);

    QMapIterator<int, QString> gameTypeIterator(allGameTypes);
    while (gameTypeIterator.hasNext()) {
        gameTypeIterator.next();
        bool enabled = gameTypeFilter.contains(gameTypeIterator.key());
        gameFilters.setGameTypeEnabled(gameTypeIterator.value(), enabled);
    }

    gameFilters.setMinPlayers(maxPlayersFilterMin);
    gameFilters.setMaxPlayers(maxPlayersFilterMax);
}

bool GamesProxyModel::filterAcceptsRow(int sourceRow, const QModelIndex & /*sourceParent*/) const
{
    return filterAcceptsRow(sourceRow);
}

bool GamesProxyModel::filterAcceptsRow(int sourceRow) const
{
    GamesModel *model = qobject_cast<GamesModel *>(sourceModel());
    if (!model)
        return false;

    const ServerInfo_Game &game = model->getGame(sourceRow);

    if (!showBuddiesOnlyGames && game.only_buddies()) {
        return false;
    }
    if (hideIgnoredUserGames && tabSupervisor->getUserListsTab()->getIgnoreList()->getUsers().contains(
                                    QString::fromStdString(game.creator_info().name()))) {
        return false;
    }
    if (!unavailableGamesVisible) {
        if (game.player_count() == game.max_players())
            return false;
        if (game.started())
            return false;
        if (!ownUserIsRegistered)
            if (game.only_registered())
                return false;
    }
    if (!showPasswordProtectedGames && game.with_password())
        return false;
    if (!gameNameFilter.isEmpty())
        if (!QString::fromStdString(game.description()).contains(gameNameFilter, Qt::CaseInsensitive))
            return false;
    if (!creatorNameFilter.isEmpty())
        if (!QString::fromStdString(game.creator_info().name()).contains(creatorNameFilter, Qt::CaseInsensitive))
            return false;

    QSet<int> gameTypes;
    for (int i = 0; i < game.game_types_size(); ++i)
        gameTypes.insert(game.game_types(i));
    if (!gameTypeFilter.isEmpty() && gameTypes.intersect(gameTypeFilter).isEmpty())
        return false;

    if ((int)game.max_players() < maxPlayersFilterMin)
        return false;
    if ((int)game.max_players() > maxPlayersFilterMax)
        return false;

    return true;
}

void GamesProxyModel::refresh()
{
    invalidateFilter();
}