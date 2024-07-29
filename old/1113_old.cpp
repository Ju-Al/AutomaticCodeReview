/***************************************************************************
 *   Copyright (C) 2009 by Andrey Afletdinov <fheroes2@gmail.com>          *
 *                                                                         *
 *   Part of the Free Heroes2 Engine:                                      *
 *   http://sourceforge.net/projects/fheroes2                              *
 *                                                                         *
 *   This program is free software; you can redistribute it and/or modify  *
 *   it under the terms of the GNU General Public License as published by  *
 *   the Free Software Foundation; either version 2 of the License, or     *
 *   (at your option) any later version.                                   *
 *                                                                         *
 *   This program is distributed in the hope that it will be useful,       *
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of        *
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the         *
 *   GNU General Public License for more details.                          *
 *                                                                         *
 *   You should have received a copy of the GNU General Public License     *
 *   along with this program; if not, write to the                         *
 *   Free Software Foundation, Inc.,                                       *
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.             *
 ***************************************************************************/

#include <algorithm>
#include <iomanip>
#include <iostream>

#include "agg.h"
#include "castle.h"
#include "game.h"
#include "game_interface.h"
#include "ground.h"
#include "heroes.h"
#include "maps.h"
#include "maps_tiles.h"
#include "monster.h"
#include "mounts.h"
#include "mp2.h"
#include "objcrck.h"
#include "objdirt.h"
#include "objdsrt.h"
#include "objgras.h"
#include "objlava.h"
#include "objmult.h"
#include "objsnow.h"
#include "objswmp.h"
#include "objtown.h"
#include "objwatr.h"
#include "objxloc.h"
#include "race.h"
#include "settings.h"
#include "spell.h"
#include "trees.h"
#include "world.h"

namespace
{
    const u8 monsterAnimationSequence[] = {0, 0, 1, 2, 1, 0, 0, 0, 3, 4, 5, 4, 3, 0, 0};

    bool contains( int base, int value )
    {
        return ( base & value ) == value;
    }
}

#ifdef WITH_DEBUG
fheroes2::Image PassableViewSurface( int passable )
{
    const u32 w = 31;
    const u32 h = 31;
    uint8_t colr = 0xBA;
    uint8_t colg = 0x5A;
    fheroes2::Image sf( w, h );
    sf.reset();

    if ( 0 == passable || Direction::CENTER == passable )
        fheroes2::DrawBorder( sf, colr );
    else if ( DIRECTION_ALL == passable )
        fheroes2::DrawBorder( sf, colg );
    else {
        for ( u32 i = 0; i < w; ++i ) {
            if ( i < 10 ) {
                fheroes2::SetPixel( sf, i, 0, ( passable & Direction::TOP_LEFT ? colg : colr ) );
                fheroes2::SetPixel( sf, i, h - 1, ( passable & Direction::BOTTOM_LEFT ? colg : colr ) );
            }
            else if ( i < 21 ) {
                fheroes2::SetPixel( sf, i, 0, ( passable & Direction::TOP ? colg : colr ) );
                fheroes2::SetPixel( sf, i, h - 1, ( passable & Direction::BOTTOM ? colg : colr ) );
            }
            else {
                fheroes2::SetPixel( sf, i, 0, ( passable & Direction::TOP_RIGHT ? colg : colr ) );
                fheroes2::SetPixel( sf, i, h - 1, ( passable & Direction::BOTTOM_RIGHT ? colg : colr ) );
            }
        }

        for ( u32 i = 0; i < h; ++i ) {
            if ( i < 10 ) {
                fheroes2::SetPixel( sf, 0, i, ( passable & Direction::TOP_LEFT ? colg : colr ) );
                fheroes2::SetPixel( sf, w - 1, i, ( passable & Direction::TOP_RIGHT ? colg : colr ) );
            }
            else if ( i < 21 ) {
                fheroes2::SetPixel( sf, 0, i, ( passable & Direction::LEFT ? colg : colr ) );
                fheroes2::SetPixel( sf, w - 1, i, ( passable & Direction::RIGHT ? colg : colr ) );
            }
            else {
                fheroes2::SetPixel( sf, 0, i, ( passable & Direction::BOTTOM_LEFT ? colg : colr ) );
                fheroes2::SetPixel( sf, w - 1, i, ( passable & Direction::BOTTOM_RIGHT ? colg : colr ) );
            }
        }
    }

    return sf;
}
#endif

Maps::TilesAddon::TilesAddon()
    : uniq( 0 )
    , level( 0 )
    , object( 0 )
    , index( 0 )
    , tmp( 0 )
{}

Maps::TilesAddon::TilesAddon( int lv, u32 gid, int obj, u32 ii )
    : uniq( gid )
    , level( lv )
    , object( obj )
    , index( ii )
    , tmp( 0 )
{}

std::string Maps::TilesAddon::String( int lvl ) const
{
    std::ostringstream os;
    os << "----------------" << lvl << "--------" << std::endl
       << "uniq            : " << uniq << std::endl
       << "tileset         : " << static_cast<int>( object ) << ", (" << ICN::GetString( MP2::GetICNObject( object ) ) << ")" << std::endl
       << "index           : " << static_cast<int>( index ) << std::endl
       << "level           : " << static_cast<int>( level ) << ", (" << static_cast<int>( level % 4 ) << ")" << std::endl
       << "tmp             : " << static_cast<int>( tmp ) << std::endl;
    return os.str();
}

Maps::TilesAddon::TilesAddon( const Maps::TilesAddon & ta )
    : uniq( ta.uniq )
    , level( ta.level )
    , object( ta.object )
    , index( ta.index )
    , tmp( ta.tmp )
{}

Maps::TilesAddon & Maps::TilesAddon::operator=( const Maps::TilesAddon & ta )
{
    level = ta.level;
    object = ta.object;
    index = ta.index;
    uniq = ta.uniq;
    tmp = ta.tmp;

    return *this;
}

bool Maps::TilesAddon::isUniq( u32 id ) const
{
    return uniq == id;
}

bool Maps::TilesAddon::isICN( int icn ) const
{
    return icn == MP2::GetICNObject( object );
}

bool Maps::TilesAddon::PredicateSortRules1( const Maps::TilesAddon & ta1, const Maps::TilesAddon & ta2 )
{
    return ( ( ta1.level % 4 ) > ( ta2.level % 4 ) );
}

bool Maps::TilesAddon::PredicateSortRules2( const Maps::TilesAddon & ta1, const Maps::TilesAddon & ta2 )
{
    return ( ( ta1.level % 4 ) < ( ta2.level % 4 ) );
}

int Maps::TilesAddon::GetLoyaltyObject( const Maps::TilesAddon & addon )
{
    switch ( MP2::GetICNObject( addon.object ) ) {
    case ICN::X_LOC1:
        if ( addon.index == 3 )
            return MP2::OBJ_ALCHEMYTOWER;
        else if ( addon.index < 3 )
            return MP2::OBJN_ALCHEMYTOWER;
        else if ( 70 == addon.index )
            return MP2::OBJ_ARENA;
        else if ( 3 < addon.index && addon.index < 72 )
            return MP2::OBJN_ARENA;
        else if ( 77 == addon.index )
            return MP2::OBJ_BARROWMOUNDS;
        else if ( 71 < addon.index && addon.index < 78 )
            return MP2::OBJN_BARROWMOUNDS;
        else if ( 94 == addon.index )
            return MP2::OBJ_EARTHALTAR;
        else if ( 77 < addon.index && addon.index < 112 )
            return MP2::OBJN_EARTHALTAR;
        else if ( 118 == addon.index )
            return MP2::OBJ_AIRALTAR;
        else if ( 111 < addon.index && addon.index < 120 )
            return MP2::OBJN_AIRALTAR;
        else if ( 127 == addon.index )
            return MP2::OBJ_FIREALTAR;
        else if ( 119 < addon.index && addon.index < 129 )
            return MP2::OBJN_FIREALTAR;
        else if ( 135 == addon.index )
            return MP2::OBJ_WATERALTAR;
        else if ( 128 < addon.index && addon.index < 137 )
            return MP2::OBJN_WATERALTAR;
        break;

    case ICN::X_LOC2:
        if ( addon.index == 4 )
            return MP2::OBJ_STABLES;
        else if ( addon.index < 4 )
            return MP2::OBJN_STABLES;
        else if ( addon.index == 9 )
            return MP2::OBJ_JAIL;
        else if ( 4 < addon.index && addon.index < 10 )
            return MP2::OBJN_JAIL;
        else if ( addon.index == 37 )
            return MP2::OBJ_MERMAID;
        else if ( 9 < addon.index && addon.index < 47 )
            return MP2::OBJN_MERMAID;
        else if ( addon.index == 101 )
            return MP2::OBJ_SIRENS;
        else if ( 46 < addon.index && addon.index < 111 )
            return MP2::OBJN_SIRENS;
        else if ( 110 < addon.index && addon.index < 136 )
            return MP2::OBJ_REEFS;
        break;

    case ICN::X_LOC3:
        if ( addon.index == 30 )
            return MP2::OBJ_HUTMAGI;
        else if ( addon.index < 32 )
            return MP2::OBJN_HUTMAGI;
        else if ( addon.index == 50 )
            return MP2::OBJ_EYEMAGI;
        else if ( 31 < addon.index && addon.index < 59 )
            return MP2::OBJN_EYEMAGI;
        // fix
        break;

    default:
        break;
    }

    return MP2::OBJ_ZERO;
}

int Maps::Tiles::GetPassable( uint32_t tileset, uint32_t index )
{
    int icn = MP2::GetICNObject( tileset );

    switch ( icn ) {
    case ICN::MTNSNOW:
    case ICN::MTNSWMP:
    case ICN::MTNLAVA:
    case ICN::MTNDSRT:
    case ICN::MTNMULT:
    case ICN::MTNGRAS:
        return ObjMnts1::GetPassable( icn, index );

    case ICN::MTNCRCK:
    case ICN::MTNDIRT:
        return ObjMnts2::GetPassable( icn, index );

    case ICN::TREJNGL:
    case ICN::TREEVIL:
    case ICN::TRESNOW:
    case ICN::TREFIR:
    case ICN::TREFALL:
    case ICN::TREDECI:
        return ObjTree::GetPassable( index );
    case ICN::OBJNSNOW:
        return ObjSnow::GetPassable( index );
    case ICN::OBJNSWMP:
        return ObjSwmp::GetPassable( index );
    case ICN::OBJNGRAS:
        return ObjGras::GetPassable( index );
    case ICN::OBJNGRA2:
        return ObjGra2::GetPassable( index );
    case ICN::OBJNCRCK:
        return ObjCrck::GetPassable( index );
    case ICN::OBJNDIRT:
        return ObjDirt::GetPassable( index );
    case ICN::OBJNDSRT:
        return ObjDsrt::GetPassable( index );
    case ICN::OBJNMUL2:
        return ObjMul2::GetPassable( index );
    case ICN::OBJNMULT:
        return ObjMult::GetPassable( index );
    case ICN::OBJNLAVA:
        return ObjLava::GetPassable( index );
    case ICN::OBJNLAV3:
        return ObjLav3::GetPassable( index );
    case ICN::OBJNLAV2:
        return ObjLav2::GetPassable( index );
    case ICN::OBJNWAT2:
        return ObjWat2::GetPassable( index );
    case ICN::OBJNWATR:
        return ObjWatr::GetPassable( index );

    case ICN::OBJNTWBA:
        return ObjTwba::GetPassable( index );
    case ICN::OBJNTOWN:
        return ObjTown::GetPassable( index );

    case ICN::X_LOC1:
        return ObjXlc1::GetPassable( index );
    case ICN::X_LOC2:
        return ObjXlc2::GetPassable( index );
    case ICN::X_LOC3:
        return ObjXlc3::GetPassable( index );

    default:
        break;
    }

    return DIRECTION_ALL;
}

int Maps::TilesAddon::GetActionObject( const Maps::TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::MTNSNOW:
    case ICN::MTNSWMP:
    case ICN::MTNLAVA:
    case ICN::MTNDSRT:
    case ICN::MTNMULT:
    case ICN::MTNGRAS:
        return ObjMnts1::GetActionObject( ta.index );

    case ICN::MTNCRCK:
    case ICN::MTNDIRT:
        return ObjMnts2::GetActionObject( ta.index );

    case ICN::TREJNGL:
    case ICN::TREEVIL:
    case ICN::TRESNOW:
    case ICN::TREFIR:
    case ICN::TREFALL:
    case ICN::TREDECI:
        return ObjTree::GetActionObject( ta.index );

    case ICN::OBJNSNOW:
        return ObjSnow::GetActionObject( ta.index );
    case ICN::OBJNSWMP:
        return ObjSwmp::GetActionObject( ta.index );
    case ICN::OBJNGRAS:
        return ObjGras::GetActionObject( ta.index );
    case ICN::OBJNGRA2:
        return ObjGra2::GetActionObject( ta.index );
    case ICN::OBJNCRCK:
        return ObjCrck::GetActionObject( ta.index );
    case ICN::OBJNDIRT:
        return ObjDirt::GetActionObject( ta.index );
    case ICN::OBJNDSRT:
        return ObjDsrt::GetActionObject( ta.index );
    case ICN::OBJNMUL2:
        return ObjMul2::GetActionObject( ta.index );
    case ICN::OBJNMULT:
        return ObjMult::GetActionObject( ta.index );
    case ICN::OBJNLAVA:
        return ObjLava::GetActionObject( ta.index );
    case ICN::OBJNLAV3:
        return ObjLav3::GetActionObject( ta.index );
    case ICN::OBJNLAV2:
        return ObjLav2::GetActionObject( ta.index );
    case ICN::OBJNWAT2:
        return ObjWat2::GetActionObject( ta.index );
    case ICN::OBJNWATR:
        return ObjWatr::GetActionObject( ta.index );

    case ICN::OBJNTWBA:
        return ObjTwba::GetActionObject( ta.index );
    case ICN::OBJNTOWN:
        return ObjTown::GetActionObject( ta.index );

    case ICN::X_LOC1:
        return ObjXlc1::GetActionObject( ta.index );
    case ICN::X_LOC2:
        return ObjXlc2::GetActionObject( ta.index );
    case ICN::X_LOC3:
        return ObjXlc3::GetActionObject( ta.index );

    default:
        break;
    }

    return MP2::OBJ_ZERO;
}

bool Maps::TilesAddon::isRoad() const
{
    switch ( MP2::GetICNObject( object ) ) {
    // from sprite road
    case ICN::ROAD:
        if ( 1 == index || 8 == index || 10 == index || 11 == index || 15 == index || 22 == index || 23 == index || 24 == index || 25 == index || 27 == index )
            return false;
        else
            return true;

    // castle and tower (gate)
    case ICN::OBJNTOWN:
        if ( 13 == index || 29 == index || 45 == index || 61 == index || 77 == index || 93 == index || 109 == index || 125 == index || 141 == index || 157 == index
             || 173 == index || 189 == index )
            return true;
        break;

        // castle lands (gate)
    case ICN::OBJNTWBA:
        if ( 7 == index || 17 == index || 27 == index || 37 == index || 47 == index || 57 == index || 67 == index || 77 == index )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::hasRoadFlag() const
{
    // This MP2 "object" is a bitfield
    // 6 bits is ICN tileset id, 1 bit isRoad flag, 1 bit hasAnimation flag
    return ( object >> 1 ) & 1;
}

bool Maps::TilesAddon::isStream( const TilesAddon & ta )
{
    return ICN::STREAM == MP2::GetICNObject( ta.object ) || ( ICN::OBJNMUL2 == MP2::GetICNObject( ta.object ) && ( ta.index < 14 ) );
}

bool Maps::TilesAddon::isRoadObject( const TilesAddon & ta )
{
    return ICN::ROAD == MP2::GetICNObject( ta.object );
}

bool Maps::TilesAddon::hasSpriteAnimation() const
{
    return object & 1;
}

bool Maps::TilesAddon::isResource( const TilesAddon & ta )
{
    // OBJNRSRC
    return ICN::OBJNRSRC == MP2::GetICNObject( ta.object ) && ( ta.index % 2 );
}

bool Maps::TilesAddon::isArtifact( const TilesAddon & ta )
{
    // OBJNARTI (skip ultimate)
    return ( ICN::OBJNARTI == MP2::GetICNObject( ta.object ) && ( ta.index > 0x10 ) && ( ta.index % 2 ) );
}

bool Maps::TilesAddon::isSkeletonFix( const TilesAddon & ta )
{
    return ( ICN::OBJNDSRT == MP2::GetICNObject( ta.object ) && ta.index == 83 );
}

bool Maps::TilesAddon::isBarrier( const TilesAddon & ta )
{
    return ICN::X_LOC3 == MP2::GetICNObject( ta.object ) && 60 <= ta.index && 102 >= ta.index && 0 == ( ta.index % 6 );
}

int Maps::TilesAddon::ColorFromBarrierSprite( const TilesAddon & ta )
{
    // 60, 66, 72, 78, 84, 90, 96, 102
    return ICN::X_LOC3 == MP2::GetICNObject( ta.object ) && 60 <= ta.index && 102 >= ta.index ? ( ( ta.index - 60 ) / 6 ) + 1 : 0;
}

int Maps::TilesAddon::ColorFromTravellerTentSprite( const TilesAddon & ta )
{
    // 110, 114, 118, 122, 126, 130, 134, 138
    return ICN::X_LOC3 == MP2::GetICNObject( ta.object ) && 110 <= ta.index && 138 >= ta.index ? ( ( ta.index - 110 ) / 4 ) + 1 : 0;
}

bool Maps::TilesAddon::isAbandoneMineSprite( const TilesAddon & ta )
{
    return ( ICN::OBJNGRAS == MP2::GetICNObject( ta.object ) && 6 == ta.index ) || ( ICN::OBJNDIRT == MP2::GetICNObject( ta.object ) && 8 == ta.index );
}

bool Maps::TilesAddon::isFlag32( const TilesAddon & ta )
{
    return ICN::FLAG32 == MP2::GetICNObject( ta.object );
}

bool Maps::TilesAddon::isX_LOC123( const TilesAddon & ta )
{
    return ( ICN::X_LOC1 == MP2::GetICNObject( ta.object ) || ICN::X_LOC2 == MP2::GetICNObject( ta.object ) || ICN::X_LOC3 == MP2::GetICNObject( ta.object ) );
}

bool Maps::TilesAddon::isShadow( const TilesAddon & ta )
{
    const int icn = MP2::GetICNObject( ta.object );

    switch ( icn ) {
    case ICN::MTNDSRT:
    case ICN::MTNGRAS:
    case ICN::MTNLAVA:
    case ICN::MTNMULT:
    case ICN::MTNSNOW:
    case ICN::MTNSWMP:
        return ObjMnts1::isShadow( ta.index );

    case ICN::MTNCRCK:
    case ICN::MTNDIRT:
        return ObjMnts2::isShadow( ta.index );

    case ICN::TREDECI:
    case ICN::TREEVIL:
    case ICN::TREFALL:
    case ICN::TREFIR:
    case ICN::TREJNGL:
    case ICN::TRESNOW:
        return ObjTree::isShadow( ta.index );

    case ICN::OBJNCRCK:
        return ObjCrck::isShadow( ta.index );
    case ICN::OBJNDIRT:
        return ObjDirt::isShadow( ta.index );
    case ICN::OBJNDSRT:
        return ObjDsrt::isShadow( ta.index );
    case ICN::OBJNGRA2:
        return ObjGra2::isShadow( ta.index );
    case ICN::OBJNGRAS:
        return ObjGras::isShadow( ta.index );
    case ICN::OBJNMUL2:
        return ObjMul2::isShadow( ta.index );
    case ICN::OBJNMULT:
        return ObjMult::isShadow( ta.index );
    case ICN::OBJNSNOW:
        return ObjSnow::isShadow( ta.index );
    case ICN::OBJNSWMP:
        return ObjSwmp::isShadow( ta.index );
    case ICN::OBJNWAT2:
        return ObjWat2::isShadow( ta.index );
    case ICN::OBJNWATR:
        return ObjWatr::isShadow( ta.index );

    case ICN::OBJNARTI:
    case ICN::OBJNRSRC:
        return 0 == ( ta.index % 2 );

    case ICN::OBJNTWRD:
        return ta.index > 31;
    case ICN::OBJNTWSH:
        return true;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isMounts( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::MTNSNOW:
    case ICN::MTNSWMP:
    case ICN::MTNLAVA:
    case ICN::MTNDSRT:
    case ICN::MTNMULT:
    case ICN::MTNGRAS:
        return !ObjMnts1::isShadow( ta.index );

    case ICN::MTNCRCK:
    case ICN::MTNDIRT:
        return !ObjMnts2::isShadow( ta.index );

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isRocs( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    // roc objects
    case ICN::OBJNSNOW:
        if ( ( ta.index > 21 && ta.index < 25 ) || ( ta.index > 25 && ta.index < 29 ) || ta.index == 30 || ta.index == 32 || ta.index == 34 || ta.index == 35
             || ( ta.index > 36 && ta.index < 40 ) )
            return true;
        break;

    case ICN::OBJNSWMP:
        if ( ta.index == 201 || ta.index == 205 || ( ta.index > 207 && ta.index < 211 ) )
            return true;
        break;

    case ICN::OBJNGRAS:
        if ( ( ta.index > 32 && ta.index < 36 ) || ta.index == 37 || ta.index == 38 || ta.index == 40 || ta.index == 41 || ta.index == 43 || ta.index == 45 )
            return true;
        break;

    case ICN::OBJNDIRT:
        if ( ta.index == 92 || ta.index == 93 || ta.index == 95 || ta.index == 98 || ta.index == 99 || ta.index == 101 || ta.index == 102 || ta.index == 104
             || ta.index == 105 )
            return true;
        break;

    case ICN::OBJNCRCK:
        if ( ta.index == 10 || ta.index == 11 || ta.index == 18 || ta.index == 19 || ta.index == 21 || ta.index == 22 || ( ta.index > 23 && ta.index < 28 )
             || ( ta.index > 28 && ta.index < 33 ) || ta.index == 34 || ta.index == 35 || ta.index == 37 || ta.index == 38 || ( ta.index > 39 && ta.index < 45 )
             || ta.index == 46 || ta.index == 47 || ta.index == 49 || ta.index == 50 || ta.index == 52 || ta.index == 53 || ta.index == 55 )
            return true;
        break;

    case ICN::OBJNWAT2:
        if ( ta.index == 0 || ta.index == 2 )
            return true;
        break;

    case ICN::OBJNWATR:
        if ( ta.index == 182 || ta.index == 183 || ( ta.index > 184 && ta.index < 188 ) )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isForests( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::TREJNGL:
    case ICN::TREEVIL:
    case ICN::TRESNOW:
    case ICN::TREFIR:
    case ICN::TREFALL:
    case ICN::TREDECI:
        return !ObjTree::isShadow( ta.index );

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isStump( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::OBJNSNOW:
        if ( ta.index == 41 || ta.index == 42 )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isDeadTrees( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::OBJNMUL2:
        if ( ta.index == 16 || ta.index == 18 || ta.index == 19 )
            return true;
        break;

    case ICN::OBJNMULT:
        if ( ta.index == 0 || ta.index == 2 || ta.index == 4 )
            return true;
        break;

    case ICN::OBJNSNOW:
        if ( ( ta.index > 50 && ta.index < 53 ) || ( ta.index > 54 && ta.index < 59 ) || ( ta.index > 59 && ta.index < 63 ) || ( ta.index > 63 && ta.index < 67 )
             || ta.index == 68 || ta.index == 69 || ta.index == 71 || ta.index == 72 || ta.index == 74 || ta.index == 75 || ta.index == 77 )
            return true;
        break;

    case ICN::OBJNSWMP:
        if ( ta.index == 161 || ta.index == 162 || ( ta.index > 163 && ta.index < 170 ) || ( ta.index > 170 && ta.index < 175 ) || ta.index == 176 || ta.index == 177 )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isCactus( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    case ICN::OBJNDSRT:
        if ( ta.index == 24 || ta.index == 26 || ta.index == 28 || ( ta.index > 29 && ta.index < 33 ) || ta.index == 34 || ta.index == 36 || ta.index == 37
             || ta.index == 39 || ta.index == 40 || ta.index == 42 || ta.index == 43 || ta.index == 45 || ta.index == 48 || ta.index == 49 || ta.index == 51
             || ta.index == 53 )
            return true;
        break;

    case ICN::OBJNCRCK:
        if ( ta.index == 14 || ta.index == 16 )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Maps::TilesAddon::isTrees( const TilesAddon & ta )
{
    switch ( MP2::GetICNObject( ta.object ) ) {
    // tree objects
    case ICN::OBJNDSRT:
        if ( ta.index == 3 || ta.index == 4 || ta.index == 6 || ta.index == 7 || ta.index == 9 || ta.index == 10 || ta.index == 12 || ta.index == 74 || ta.index == 76 )
            return true;
        break;

    case ICN::OBJNDIRT:
        if ( ta.index == 115 || ta.index == 118 || ta.index == 120 || ta.index == 123 || ta.index == 125 || ta.index == 127 )
            return true;
        break;

    case ICN::OBJNGRAS:
        if ( ta.index == 80 || ( ta.index > 82 && ta.index < 86 ) || ta.index == 87 || ( ta.index > 88 && ta.index < 92 ) )
            return true;
        break;

    default:
        break;
    }

    return false;
}

void Maps::Tiles::UpdateAbandoneMineLeftSprite( uint8_t & tileset, uint8_t & index, int resource )
{
    if ( ICN::OBJNGRAS == MP2::GetICNObject( tileset ) && 6 == index ) {
        tileset = 128; // MTNGRAS
        index = 82;
    }
    else if ( ICN::OBJNDIRT == MP2::GetICNObject( tileset ) && 8 == index ) {
        tileset = 104; // MTNDIRT
        index = 112;
    }
    else if ( ICN::EXTRAOVR == MP2::GetICNObject( tileset ) && 5 == index ) {
        switch ( resource ) {
        case Resource::ORE:
            index = 0;
            break;
        case Resource::SULFUR:
            index = 1;
            break;
        case Resource::CRYSTAL:
            index = 2;
            break;
        case Resource::GEMS:
            index = 3;
            break;
        case Resource::GOLD:
            index = 4;
            break;
        default:
            break;
        }
    }
}

void Maps::Tiles::UpdateAbandoneMineRightSprite( uint8_t & tileset, uint8_t & index )
{
    if ( ICN::OBJNDIRT == MP2::GetICNObject( tileset ) && index == 9 ) {
        tileset = 104;
        index = 113;
    }
    else if ( ICN::OBJNGRAS == MP2::GetICNObject( tileset ) && index == 7 ) {
        tileset = 128;
        index = 83;
    }
}

std::pair<int, int> Maps::Tiles::ColorRaceFromHeroSprite( uint32_t heroSpriteIndex )
{
    std::pair<int, int> res;

    if ( 7 > heroSpriteIndex )
        res.first = Color::BLUE;
    else if ( 14 > heroSpriteIndex )
        res.first = Color::GREEN;
    else if ( 21 > heroSpriteIndex )
        res.first = Color::RED;
    else if ( 28 > heroSpriteIndex )
        res.first = Color::YELLOW;
    else if ( 35 > heroSpriteIndex )
        res.first = Color::ORANGE;
    else
        res.first = Color::PURPLE;

    switch ( heroSpriteIndex % 7 ) {
    case 0:
        res.second = Race::KNGT;
        break;
    case 1:
        res.second = Race::BARB;
        break;
    case 2:
        res.second = Race::SORC;
        break;
    case 3:
        res.second = Race::WRLK;
        break;
    case 4:
        res.second = Race::WZRD;
        break;
    case 5:
        res.second = Race::NECR;
        break;
    case 6:
        res.second = Race::RAND;
        break;
    }

    return res;
}

bool Maps::TilesAddon::ForceLevel1( const TilesAddon & ta )
{
    // broken ship: left roc
    if ( ICN::OBJNWAT2 == MP2::GetICNObject( ta.object ) && ta.index == 11 )
        return true;

    return false;
}

bool Maps::TilesAddon::ForceLevel2( const TilesAddon & ta )
{
    return false;
}

/* Maps::Addons */
void Maps::Addons::Remove( u32 uniq )
{
    /*
    erase(std::remove_if(begin(), end(),
        std::bind2nd(std::mem_fun_ref(&Maps::TilesAddon::isUniq), uniq)),
        end());
    */
    remove_if( std::bind2nd( std::mem_fun_ref( &Maps::TilesAddon::isUniq ), uniq ) );
}

u32 PackTileSpriteIndex( u32 index, u32 shape ) /* index max: 0x3FFF, shape value: 0, 1, 2, 3 */
{
    return ( shape << 14 ) | ( 0x3FFF & index );
}

/* Maps::Tiles */
Maps::Tiles::Tiles()
    : maps_index( 0 )
    , pack_sprite_index( 0 )
    , tilePassable( DIRECTION_ALL )
    , uniq( 0 )
    , mp2_object( 0 )
    , objectTileset( 0 )
    , objectIndex( 255 )
    , fog_colors( Color::ALL )
    , quantity1( 0 )
    , quantity2( 0 )
    , quantity3( 0 )
#ifdef WITH_DEBUG
    , impassableTileRule( 0 )
#endif
{}

void Maps::Tiles::Init( s32 index, const MP2::mp2tile_t & mp2 )
{
    tilePassable = DIRECTION_ALL;
    quantity1 = mp2.quantity1;
    quantity2 = mp2.quantity2;
    quantity3 = 0;
    fog_colors = Color::ALL;

    SetTile( mp2.tileIndex, mp2.flags );
    SetIndex( index );
    SetObject( mp2.mapObject );

    addons_level1.clear();
    addons_level2.clear();

    // those bitfields are set by map editor regardless if map object is there
    tileIsRoad = ( mp2.objectName1 >> 1 ) & 1;

    if ( mp2.mapObject == MP2::OBJ_ZERO || ( quantity1 >> 1 ) & 1 ) {
        AddonsPushLevel1( mp2 );
    }
    else {
        objectTileset = mp2.objectName1;
        objectIndex = mp2.indexName1;
        uniq = mp2.editorObjectLink;
    }
    AddonsPushLevel2( mp2 );
}

void Maps::Tiles::SetIndex( int index )
{
    maps_index = index;
}

int Maps::Tiles::GetQuantity3( void ) const
{
    return quantity3;
}

void Maps::Tiles::SetQuantity3( int mod )
void Maps::Tiles::SetQuantity3( int value )
{
    quantity3 = value;
}

Heroes * Maps::Tiles::GetHeroes( void ) const
{
    return MP2::OBJ_HEROES == mp2_object && heroID ? world.GetHeroes( heroID - 1 ) : NULL;
}

void Maps::Tiles::SetHeroes( Heroes * hero )
{
    if ( hero ) {
        hero->SetMapsObject( mp2_object );
        heroID = hero->GetID() + 1;
        SetObject( MP2::OBJ_HEROES );
    }
    else {
        hero = GetHeroes();

        if ( hero ) {
            SetObject( hero->GetMapsObject() );
            hero->SetMapsObject( MP2::OBJ_ZERO );
        }
        else
            SetObject( MP2::OBJ_ZERO );

        heroID = 0;
    }
}

Point Maps::Tiles::GetCenter( void ) const
{
    return Maps::GetPoint( GetIndex() );
}

s32 Maps::Tiles::GetIndex( void ) const
{
    return maps_index;
}

int Maps::Tiles::GetObject( bool ignoreObjectUnderHero /* true */ ) const
{
    if ( !ignoreObjectUnderHero && MP2::OBJ_HEROES == mp2_object ) {
        const Heroes * hero = GetHeroes();
        return hero ? hero->GetMapsObject() : MP2::OBJ_ZERO;
    }

    return mp2_object;
}

void Maps::Tiles::SetObject( int object )
{
    mp2_object = object;
}

void Maps::Tiles::SetTile( u32 sprite_index, u32 shape )
{
    pack_sprite_index = PackTileSpriteIndex( sprite_index, shape );
}

u32 Maps::Tiles::TileSpriteIndex( void ) const
{
    return pack_sprite_index & 0x3FFF;
}

u32 Maps::Tiles::TileSpriteShape( void ) const
{
    return pack_sprite_index >> 14;
}

const fheroes2::Image & Maps::Tiles::GetTileSurface( void ) const
{
    return fheroes2::AGG::GetTIL( TIL::GROUND32, TileSpriteIndex(), TileSpriteShape() );
}

bool isMountsRocs( const Maps::TilesAddon & ta )
{
    return Maps::TilesAddon::isMounts( ta ) || Maps::TilesAddon::isRocs( ta );
}

bool isForestsTrees( const Maps::TilesAddon & ta )
{
    return Maps::TilesAddon::isForests( ta ) || Maps::TilesAddon::isTrees( ta ) || Maps::TilesAddon::isCactus( ta );
}

// Return true if tile will be impassable if another object overlays it. Edge case when dealing with mountain and forest tiles.
bool isImpassableIfOverlayed( uint8_t objectTileset, uint8_t icnIndex )
{
    // Unfortunately can't be determined by the object ID, have to check the tiles
    switch ( MP2::GetICNObject( objectTileset ) ) {
    case ICN::MTNSNOW:
    case ICN::MTNSWMP:
    case ICN::MTNLAVA:
    case ICN::MTNDSRT:
    case ICN::MTNMULT:
    case ICN::MTNGRAS:
        return !ObjMnts1::isShadow( icnIndex );

    case ICN::MTNCRCK:
    case ICN::MTNDIRT:
        return !ObjMnts2::isShadow( icnIndex );

    case ICN::TREJNGL:
    case ICN::TREEVIL:
    case ICN::TRESNOW:
    case ICN::TREFIR:
    case ICN::TREFALL:
    case ICN::TREDECI:
        return !ObjTree::isShadow( icnIndex );

    case ICN::OBJNSNOW:
        if ( ( icnIndex > 21 && icnIndex < 25 ) || ( icnIndex > 25 && icnIndex < 29 ) || icnIndex == 30 || icnIndex == 32 || icnIndex == 34 || icnIndex == 35
             || ( icnIndex > 36 && icnIndex < 40 ) )
            return true;
        break;

    case ICN::OBJNSWMP:
        if ( icnIndex == 201 || icnIndex == 205 || ( icnIndex > 207 && icnIndex < 211 ) )
            return true;
        break;

    case ICN::OBJNGRAS:
        if ( ( icnIndex > 32 && icnIndex < 36 ) || icnIndex == 37 || icnIndex == 38 || icnIndex == 40 || icnIndex == 41 || icnIndex == 43 || icnIndex == 45
             || icnIndex == 80 || ( icnIndex > 82 && icnIndex < 86 ) || icnIndex == 87 || ( icnIndex > 88 && icnIndex < 92 ) )
            return true;
        break;

    case ICN::OBJNDIRT:
        if ( icnIndex == 92 || icnIndex == 93 || icnIndex == 95 || icnIndex == 98 || icnIndex == 99 || icnIndex == 101 || icnIndex == 102 || icnIndex == 104
             || icnIndex == 105 || icnIndex == 115 || icnIndex == 118 || icnIndex == 120 || icnIndex == 123 || icnIndex == 125 || icnIndex == 127 )
            return true;
        break;

    case ICN::OBJNCRCK:
        if ( icnIndex == 10 || icnIndex == 11 || icnIndex == 14 || icnIndex == 16 || icnIndex == 18 || icnIndex == 19 || icnIndex == 21 || icnIndex == 22
             || ( icnIndex > 23 && icnIndex < 28 ) || ( icnIndex > 28 && icnIndex < 33 ) || icnIndex == 34 || icnIndex == 35 || icnIndex == 37 || icnIndex == 38
             || ( icnIndex > 39 && icnIndex < 45 ) || icnIndex == 46 || icnIndex == 47 || icnIndex == 49 || icnIndex == 50 || icnIndex == 52 || icnIndex == 53
             || icnIndex == 55 )
            return true;
        break;

    case ICN::OBJNWAT2:
        if ( icnIndex == 0 || icnIndex == 2 )
            return true;
        break;

    case ICN::OBJNWATR:
        if ( icnIndex == 182 || icnIndex == 183 || ( icnIndex > 184 && icnIndex < 188 ) )
            return true;
        break;

    case ICN::OBJNDSRT:
        if ( icnIndex == 3 || icnIndex == 4 || icnIndex == 6 || icnIndex == 7 || icnIndex == 9 || icnIndex == 10 || icnIndex == 12 || icnIndex == 24 || icnIndex == 26
             || icnIndex == 28 || ( icnIndex > 29 && icnIndex < 33 ) || icnIndex == 34 || icnIndex == 36 || icnIndex == 37 || icnIndex == 39 || icnIndex == 40
             || icnIndex == 42 || icnIndex == 43 || icnIndex == 45 || icnIndex == 48 || icnIndex == 49 || icnIndex == 51 || icnIndex == 53 || icnIndex == 74
             || icnIndex == 76 )
            return true;
        break;

    default:
        break;
    }

    return false;
}

bool Exclude4LongObject( const Maps::TilesAddon & ta )
{
    return Maps::TilesAddon::isStream( ta ) || Maps::TilesAddon::isRoadObject( ta ) || Maps::TilesAddon::isShadow( ta );
}

bool HaveLongObjectUniq( const Maps::Addons & level, u32 uid )
{
    for ( Maps::Addons::const_iterator it = level.begin(); it != level.end(); ++it )
        if ( !Exclude4LongObject( *it ) && ( *it ).isUniq( uid ) )
            return true;
    return false;
}

bool Maps::Tiles::isLongObject( int direction )
{
    if ( Maps::isValidDirection( GetIndex(), direction ) ) {
        Tiles & tile = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), direction ) );

        for ( Addons::const_iterator it = addons_level1.begin(); it != addons_level1.end(); ++it )
            if ( !Exclude4LongObject( *it )
                 && ( HaveLongObjectUniq( tile.addons_level1, ( *it ).uniq )
                      || ( !Maps::TilesAddon::isTrees( *it ) && HaveLongObjectUniq( tile.addons_level2, ( *it ).uniq ) ) ) )
                return true;
    }
    return false;
}

void Maps::Tiles::UpdatePassable( void )
{
    tilePassable = DIRECTION_ALL;
#ifdef WITH_DEBUG
    impassableTileRule = 0;
#endif

    const int obj = GetObject( false );
    const bool emptyobj = MP2::OBJ_ZERO == obj || MP2::OBJ_COAST == obj || MP2::OBJ_EVENT == obj;

    if ( MP2::isActionObject( obj, isWater() ) ) {
        tilePassable = MP2::GetObjectDirect( obj );
#ifdef WITH_DEBUG
        if ( tilePassable == 0 )
            impassableTileRule = 1;
#endif
        return;
    }

    // on ground
    if ( MP2::OBJ_HEROES != mp2_object && !isWater() ) {
        bool hasRocksOrTrees = isImpassableIfOverlayed( objectTileset, objectIndex );
        bool mounts2 = addons_level2.end() != std::find_if( addons_level2.begin(), addons_level2.end(), isMountsRocs );
        bool trees2 = addons_level2.end() != std::find_if( addons_level2.begin(), addons_level2.end(), isForestsTrees );

        // fix coast passable
        if ( tilePassable && !emptyobj && Maps::TileIsCoast( GetIndex(), Direction::TOP | Direction::BOTTOM | Direction::LEFT | Direction::RIGHT ) && !isShadow() ) {
            tilePassable = 0;
#ifdef WITH_DEBUG
            impassableTileRule = 2;
#endif
        }

        // fix mountain layer
        if ( tilePassable && hasRocksOrTrees && ( mounts2 || trees2 ) ) {
            tilePassable = 0;
#ifdef WITH_DEBUG
            impassableTileRule = 3;
#endif
        }

        // town twba
        if ( tilePassable && FindAddonICN( ICN::OBJNTWBA, 1 ) && ( mounts2 || trees2 ) ) {
            tilePassable = 0;
#ifdef WITH_DEBUG
            impassableTileRule = 5;
#endif
        }

        if ( Maps::isValidDirection( GetIndex(), Direction::TOP ) ) {
            Tiles & top = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::TOP ) );
            // fix: rocs on water
            if ( top.isWater() && top.tilePassable && !( Direction::TOP & top.tilePassable ) ) {
                top.tilePassable = 0;
#ifdef WITH_DEBUG
                top.impassableTileRule = 6;
#endif
            }
        }
    }

    // fix bottom border: disable passable for all no action objects
    if ( tilePassable && !Maps::isValidDirection( GetIndex(), Direction::BOTTOM ) && !emptyobj && !MP2::isActionObject( obj, isWater() ) ) {
        tilePassable = 0;
#ifdef WITH_DEBUG
        impassableTileRule = 7;
#endif
    }

    // check all sprite (level 1)
    tilePassable &= Tiles::GetPassable( objectTileset, objectIndex );
    for ( Addons::const_iterator it = addons_level1.begin(); it != addons_level1.end(); ++it ) {
        if ( tilePassable ) {
            tilePassable &= Tiles::GetPassable( it->object, it->index );

#ifdef WITH_DEBUG
            if ( 0 == tilePassable )
                impassableTileRule = 8;
#endif
        }
    }

    // fix top passable
    if ( Maps::isValidDirection( GetIndex(), Direction::TOP ) ) {
        Tiles & top = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::TOP ) );

        if ( isWater() == top.isWater() && isImpassableIfOverlayed( top.objectTileset, top.objectIndex ) && !MP2::isActionObject( top.GetObject( false ), isWater() )
             && ( tilePassable && !( tilePassable & DIRECTION_TOP_ROW ) ) && !( top.tilePassable & DIRECTION_TOP_ROW ) ) {
            top.tilePassable = 0;
#ifdef WITH_DEBUG
            top.impassableTileRule = 9;
#endif
        }
    }

    // fix corners
    if ( Maps::isValidDirection( GetIndex(), Direction::LEFT ) ) {
        Tiles & left = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::LEFT ) );

        // left corner
        if ( left.tilePassable && isLongObject( Direction::TOP ) && !( ( Direction::TOP | Direction::TOP_LEFT ) & tilePassable )
             && ( Direction::TOP_RIGHT & left.tilePassable ) ) {
            left.tilePassable &= ~Direction::TOP_RIGHT;
        }
        else
            // right corner
            if ( tilePassable && left.isLongObject( Direction::TOP ) && !( ( Direction::TOP | Direction::TOP_RIGHT ) & left.tilePassable )
                 && ( Direction::TOP_LEFT & tilePassable ) ) {
            tilePassable &= ~Direction::TOP_LEFT;
        }
    }
}

u32 Maps::Tiles::GetObjectUID() const
{
    return uniq;
}

// Get Tile metadata field #1 (used for things like monster count of resource amount)
int Maps::Tiles::GetQuantity1( void ) const
{
    return quantity1;
}

// Get Tile metadata field #2 (used for things like animations or resource type )
int Maps::Tiles::GetQuantity2( void ) const
{
    return quantity2;
}

int Maps::Tiles::GetPassable( void ) const
{
    return tilePassable;
}

void Maps::Tiles::AddonsPushLevel1( const MP2::mp2tile_t & mt )
{
    if ( mt.objectName1 && mt.indexName1 < 0xFF )
        AddonsPushLevel1( TilesAddon( 0, mt.editorObjectLink, mt.objectName1, mt.indexName1 ) );

    // MP2 "objectName" is a bitfield
    // 6 bits is ICN tileset id, 1 bit isRoad flag, 1 bit hasAnimation flag
    if ( ( mt.objectName1 >> 1 ) & 1 )
        tileIsRoad = true;
}

void Maps::Tiles::AddonsPushLevel1( const MP2::mp2addon_t & ma )
{
    if ( ma.objectNameN1 && ma.indexNameN1 < 0xFF )
        AddonsPushLevel1( TilesAddon( ma.quantityN, ma.editorObjectLink, ma.objectNameN1, ma.indexNameN1 ) );
}

void Maps::Tiles::AddonsPushLevel1( const TilesAddon & ta )
{
    if ( TilesAddon::ForceLevel2( ta ) )
        addons_level2.push_back( ta );
    else
        addons_level1.push_back( ta );
}

void Maps::Tiles::AddonsPushLevel2( const MP2::mp2tile_t & mt )
{
    if ( mt.objectName2 && mt.indexName2 < 0xFF )
        AddonsPushLevel2( TilesAddon( 0, mt.editorObjectOverlay, mt.objectName2, mt.indexName2 ) );
}

void Maps::Tiles::AddonsPushLevel2( const MP2::mp2addon_t & ma )
{
    if ( ma.objectNameN2 && ma.indexNameN2 < 0xFF )
        AddonsPushLevel2( TilesAddon( ma.quantityN, ma.editorObjectOverlay, ma.objectNameN2, ma.indexNameN2 ) );
}

void Maps::Tiles::AddonsPushLevel2( const TilesAddon & ta )
{
    if ( TilesAddon::ForceLevel1( ta ) )
        addons_level1.push_back( ta );
    else
        addons_level2.push_back( ta );
}

void Maps::Tiles::AddonsSort( void )
{
    Addons::iterator lastObject = addons_level1.end();
    for ( Addons::iterator it = addons_level1.begin(); it != addons_level1.end(); ++it ) {
        // force monster object on top (fix old saves)
        if ( mp2_object == MP2::OBJ_MONSTER && MP2::GetICNObject( it->object ) == ICN::MONS32 ) {
            lastObject = it;
            break;
        }
        // Level is actually a bitfield, but if 0 then it's an actual "full object"
        // To compare: level 2 is shadow, level 3 is roads/river
        else if ( it->level % 4 == 0 ) {
            lastObject = it;
        }
    }

    // Re-order tiles (Fixing mistakes of original maps)
    if ( lastObject != addons_level1.end() ) {
        if ( objectTileset != 0 && objectIndex < 255 )
            addons_level1.push_front( TilesAddon( 0, uniq, objectTileset, objectIndex ) );

        uniq = lastObject->uniq;
        objectTileset = lastObject->object;
        objectIndex = lastObject->index;
        addons_level1.erase( lastObject );
    }

    // FIXME: check if sort is still needed
    // if ( !addons_level1.empty() )
    //    addons_level1.sort( TilesAddon::PredicateSortRules1 );
    // if ( !addons_level2.empty() )
    //    addons_level2.sort( TilesAddon::PredicateSortRules2 );
}

int Maps::Tiles::GetGround( void ) const
{
    const u32 index = TileSpriteIndex();

    // list grounds from GROUND32.TIL
    if ( 30 > index )
        return Maps::Ground::WATER;
    else if ( 92 > index )
        return Maps::Ground::GRASS;
    else if ( 146 > index )
        return Maps::Ground::SNOW;
    else if ( 208 > index )
        return Maps::Ground::SWAMP;
    else if ( 262 > index )
        return Maps::Ground::LAVA;
    else if ( 321 > index )
        return Maps::Ground::DESERT;
    else if ( 361 > index )
        return Maps::Ground::DIRT;
    else if ( 415 > index )
        return Maps::Ground::WASTELAND;

    return Maps::Ground::BEACH;
}

bool Maps::Tiles::isWater( void ) const
{
    return 30 > TileSpriteIndex();
}

void Maps::Tiles::RedrawTile( fheroes2::Image & dst ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( area.GetVisibleTileROI() & mp )
        area.BlitOnTile( dst, GetTileSurface(), 0, 0, mp );
}

void Maps::Tiles::RedrawEmptyTile( fheroes2::Image & dst, const Point & mp )
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();

    if ( area.GetVisibleTileROI() & mp ) {
        if ( mp.y == -1 && mp.x >= 0 && mp.x < world.w() ) { // top first row
            area.BlitOnTile( dst, fheroes2::AGG::GetTIL( TIL::STON, 20 + ( mp.x % 4 ), 0 ), 0, 0, mp );
        }
        else if ( mp.x == world.w() && mp.y >= 0 && mp.y < world.h() ) { // right first row
            area.BlitOnTile( dst, fheroes2::AGG::GetTIL( TIL::STON, 24 + ( mp.y % 4 ), 0 ), 0, 0, mp );
        }
        else if ( mp.y == world.h() && mp.x >= 0 && mp.x < world.w() ) { // bottom first row
            area.BlitOnTile( dst, fheroes2::AGG::GetTIL( TIL::STON, 28 + ( mp.x % 4 ), 0 ), 0, 0, mp );
        }
        else if ( mp.x == -1 && mp.y >= 0 && mp.y < world.h() ) { // left first row
            area.BlitOnTile( dst, fheroes2::AGG::GetTIL( TIL::STON, 32 + ( mp.y % 4 ), 0 ), 0, 0, mp );
        }
        else {
            area.BlitOnTile( dst, fheroes2::AGG::GetTIL( TIL::STON, ( abs( mp.y ) % 4 ) * 4 + abs( mp.x ) % 4, 0 ), 0, 0, mp );
        }
    }
}

void Maps::Tiles::RedrawAddon( fheroes2::Image & dst, const Addons & addon, bool skipObjs ) const
{
    Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( ( area.GetVisibleTileROI() & mp ) && !addon.empty() ) {
        for ( Addons::const_iterator it = addon.begin(); it != addon.end(); ++it ) {
            // skip
            if ( skipObjs && MP2::isRemoveObject( GetObject() ) )
                continue;

            const u8 & index = ( *it ).index;
            const int icn = MP2::GetICNObject( ( *it ).object );

            if ( ICN::UNKNOWN != icn && ICN::MINIHERO != icn && ICN::MONS32 != icn ) {
                const fheroes2::Sprite & sprite = fheroes2::AGG::GetICN( icn, index );
                area.BlitOnTile( dst, sprite, sprite.x(), sprite.y(), mp );

                // possible animation
                uint32_t animationIndex = ICN::AnimationFrame( icn, index, Game::MapsAnimationFrame(), quantity2 );
                if ( animationIndex ) {
                    area.BlitOnTile( dst, fheroes2::AGG::GetICN( icn, animationIndex ), mp );
                }
            }
        }
    }
}

void Maps::Tiles::RedrawBottom( fheroes2::Image & dst, bool skipObjs ) const
{
    RedrawAddon( dst, addons_level1, skipObjs );
}

void Maps::Tiles::RedrawPassable( fheroes2::Image & dst ) const
{
#ifdef WITH_DEBUG
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( area.GetVisibleTileROI() & mp ) {
        if ( 0 == tilePassable || DIRECTION_ALL != tilePassable ) {
            fheroes2::Image sf = PassableViewSurface( tilePassable );

            if ( impassableTileRule ) {
                Text text( GetString( impassableTileRule ), Font::SMALL );
                text.Blit( 13, 13, sf );
            }

            area.BlitOnTile( dst, sf, 0, 0, mp );
        }
    }
#endif
}

void Maps::Tiles::RedrawObjects( fheroes2::Image & dst ) const
{
    switch ( GetObject() ) {
    // boat
    case MP2::OBJ_BOAT:
        RedrawBoat( dst );
        break;
    // monster
    case MP2::OBJ_MONSTER:
        RedrawMonster( dst );
        break;
    //
    default: {
        const int icn = MP2::GetICNObject( objectTileset );

        if ( ICN::UNKNOWN != icn ) {
            const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
            const Point mp = Maps::GetPoint( GetIndex() );

            const fheroes2::Sprite & sprite = fheroes2::AGG::GetICN( icn, objectIndex );
            area.BlitOnTile( dst, sprite, sprite.x(), sprite.y(), mp );

            // possible animation
            uint32_t animationIndex = ICN::AnimationFrame( icn, objectIndex, Game::MapsAnimationFrame(), quantity2 );
            if ( animationIndex ) {
                const fheroes2::Sprite & animationSprite = fheroes2::AGG::GetICN( icn, animationIndex );

                area.BlitOnTile( dst, animationSprite, mp );
            }
        }
    } break;
    }
}

void Maps::Tiles::RedrawMonster( fheroes2::Image & dst ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( !( area.GetVisibleTileROI() & mp ) )
        return;

    const Monster & monster = QuantityMonster();
    const std::pair<int, int> spriteIndicies = GetMonsterSpriteIndices( *this, monster.GetSpriteIndex() );

    const fheroes2::Sprite & sprite = fheroes2::AGG::GetICN( ICN::MINIMON, spriteIndicies.first );
    area.BlitOnTile( dst, sprite, sprite.x() + 16, sprite.y() + TILEWIDTH, mp );

    if ( spriteIndicies.second != -1 ) {
        const fheroes2::Sprite & animatedSprite = fheroes2::AGG::GetICN( ICN::MINIMON, spriteIndicies.second );
        area.BlitOnTile( dst, animatedSprite, animatedSprite.x() + 16, animatedSprite.y() + TILEWIDTH, mp );
    }
}

void Maps::Tiles::RedrawBoat( fheroes2::Image & dst ) const
{
    const Point mp = Maps::GetPoint( GetIndex() );
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();

    if ( area.GetVisibleTileROI() & mp ) {
        // FIXME: restore direction from Maps::Tiles
        const fheroes2::Sprite & sprite = fheroes2::AGG::GetICN( ICN::BOAT32, 18 );
        area.BlitOnTile( dst, sprite, sprite.x(), TILEWIDTH + sprite.y(), mp );
    }
}

bool SkipRedrawTileBottom4Hero( uint8_t tileset, uint8_t icnIndex, int passable )
{
    switch ( MP2::GetICNObject( tileset ) ) {
    case ICN::UNKNOWN:
    case ICN::MINIHERO:
    case ICN::MONS32:
        return true;

    // whirlpool
    case ICN::OBJNWATR:
        return icnIndex >= 202 && icnIndex <= 225;

    // river delta
    case ICN::OBJNMUL2:
        return icnIndex < 14;

    case ICN::OBJNTWSH:
    case ICN::OBJNTWBA:
    case ICN::ROAD:
    case ICN::STREAM:
        return true;

    case ICN::OBJNCRCK:
        return ( icnIndex == 58 || icnIndex == 59 || icnIndex == 64 || icnIndex == 65 || icnIndex == 188 || icnIndex == 189 || ( passable & DIRECTION_TOP_ROW ) );

    case ICN::OBJNDIRT:
    case ICN::OBJNDSRT:
    case ICN::OBJNGRA2:
    case ICN::OBJNGRAS:
    case ICN::OBJNLAVA:
    case ICN::OBJNSNOW:
    case ICN::OBJNSWMP:
        return ( passable & DIRECTION_TOP_ROW );

    default:
        break;
    }

    return false;
}

void Maps::Tiles::RedrawBottom4Hero( fheroes2::Image & dst ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( ( area.GetVisibleTileROI() & mp ) ) {
        for ( Addons::const_iterator it = addons_level1.begin(); it != addons_level1.end(); ++it ) {
            if ( !SkipRedrawTileBottom4Hero( it->object, it->index, tilePassable ) ) {
                const u8 & object = ( *it ).object;
                const u8 & index = ( *it ).index;
                const int icn = MP2::GetICNObject( object );

                area.BlitOnTile( dst, fheroes2::AGG::GetICN( icn, index ), mp );

                // possible anime
                if ( it->object & 1 ) {
                    area.BlitOnTile( dst, fheroes2::AGG::GetICN( icn, ICN::AnimationFrame( icn, index, Game::MapsAnimationFrame(), quantity2 ) ), mp );
                }
            }
        }

        if ( !SkipRedrawTileBottom4Hero( objectTileset, objectIndex, tilePassable ) ) {
            RedrawObjects( dst );
        }
    }
}

void Maps::Tiles::RedrawTop( fheroes2::Image & dst, bool skipObjs ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( !( area.GetVisibleTileROI() & mp ) )
        return;

    // animate objects
    if ( MP2::OBJ_ABANDONEDMINE == GetObject() ) {
        area.BlitOnTile( dst, fheroes2::AGG::GetICN( ICN::OBJNHAUN, Game::MapsAnimationFrame() % 15 ), mp );
    }
    else if ( MP2::OBJ_MINES == GetObject() ) {
        const uint8_t spellID = GetQuantity3();
        if ( spellID == Spell::HAUNT ) {
            area.BlitOnTile( dst, fheroes2::AGG::GetICN( ICN::OBJNHAUN, Game::MapsAnimationFrame() % 15 ), mp );
        }
        else if ( spellID >= Spell::SETEGUARDIAN && spellID <= Spell::SETWGUARDIAN ) {
            area.BlitOnTile( dst, fheroes2::AGG::GetICN( ICN::MONS32, Monster( Spell( spellID ) ).GetSpriteIndex() ), TILEWIDTH, 0, mp );
        }
    }

    RedrawAddon( dst, addons_level2, skipObjs );
}

void Maps::Tiles::RedrawTop4Hero( fheroes2::Image & dst, bool skip_ground ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    if ( ( area.GetVisibleTileROI() & mp ) && !addons_level2.empty() ) {
        for ( Addons::const_iterator it = addons_level2.begin(); it != addons_level2.end(); ++it ) {
            if ( skip_ground && MP2::isGroundObject( ( *it ).object ) )
                continue;

            const u8 & object = ( *it ).object;
            const u8 & index = ( *it ).index;
            const int icn = MP2::GetICNObject( object );

            if ( ICN::HighlyObjectSprite( icn, index ) ) {
                area.BlitOnTile( dst, fheroes2::AGG::GetICN( icn, index ), mp );

                // possible anime
                if ( object & 1 ) {
                    area.BlitOnTile( dst, fheroes2::AGG::GetICN( icn, ICN::AnimationFrame( icn, index, Game::MapsAnimationFrame() ) ), mp );
                }
            }
        }
    }
}

Maps::TilesAddon * Maps::Tiles::FindAddonICN( int icn, int level, int index )
{
    if ( level == 1 || level == -1 ) {
        for ( Addons::iterator it = addons_level1.begin(); it != addons_level1.end(); ++it ) {
            if ( MP2::GetICNObject( it->object ) == icn && ( index == -1 || index == it->index ) ) {
                return &( *it );
            }
        }
    }
    if ( level == 2 || level == -1 ) {
        for ( Addons::iterator it = addons_level2.begin(); it != addons_level2.end(); ++it ) {
            if ( MP2::GetICNObject( it->object ) == icn && ( index == -1 || index == it->index ) ) {
                return &( *it );
            }
        }
    }
    return NULL;
}

Maps::TilesAddon * Maps::Tiles::FindAddonLevel1( u32 uniq1 )
{
    Addons::iterator it = std::find_if( addons_level1.begin(), addons_level1.end(), std::bind2nd( std::mem_fun_ref( &Maps::TilesAddon::isUniq ), uniq1 ) );

    return it != addons_level1.end() ? &( *it ) : NULL;
}

Maps::TilesAddon * Maps::Tiles::FindAddonLevel2( u32 uniq2 )
{
    Addons::iterator it = std::find_if( addons_level2.begin(), addons_level2.end(), std::bind2nd( std::mem_fun_ref( &Maps::TilesAddon::isUniq ), uniq2 ) );

    return it != addons_level2.end() ? &( *it ) : NULL;
}

std::string Maps::Tiles::String( void ) const
{
    std::ostringstream os;

    os << "----------------:--------" << std::endl
       << "maps index      : " << GetIndex() << ", " << GetString( GetCenter() ) << std::endl
       << "id              : " << uniq << std::endl
       << "mp2 object      : " << static_cast<int>( GetObject() ) << ", (" << MP2::StringObject( GetObject() ) << ")" << std::endl
       << "tileset         : " << static_cast<int>( objectTileset ) << ", (" << ICN::GetString( MP2::GetICNObject( objectTileset ) ) << ")" << std::endl
       << "object index    : " << static_cast<int>( objectIndex ) << std::endl
       << "object animated : " << static_cast<int>( hasSpriteAnimation() ) << std::endl
       << "ground          : " << Ground::String( GetGround() ) << ", (isRoad: " << tileIsRoad << ")" << std::endl
       << "passable        : " << ( tilePassable ? Direction::String( tilePassable ) : "false" );
#ifdef WITH_DEBUG
    if ( impassableTileRule )
        os << ", disable(" << static_cast<int>( impassableTileRule ) << ")";
#endif
    os << std::endl
       << "quantity 1      : " << static_cast<int>( quantity1 ) << std::endl
       << "quantity 2      : " << static_cast<int>( quantity2 ) << std::endl
       << "quantity 3      : " << static_cast<int>( GetQuantity3() ) << std::endl;

    for ( Addons::const_iterator it = addons_level1.begin(); it != addons_level1.end(); ++it )
        os << ( *it ).String( 1 );

    for ( Addons::const_iterator it = addons_level2.begin(); it != addons_level2.end(); ++it )
        os << ( *it ).String( 2 );

    os << "----------------I--------" << std::endl;

    // extra obj info
    switch ( GetObject() ) {
        // dwelling
    case MP2::OBJ_RUINS:
    case MP2::OBJ_TREECITY:
    case MP2::OBJ_WAGONCAMP:
    case MP2::OBJ_DESERTTENT:
    case MP2::OBJ_TROLLBRIDGE:
    case MP2::OBJ_DRAGONCITY:
    case MP2::OBJ_CITYDEAD:
        //
    case MP2::OBJ_WATCHTOWER:
    case MP2::OBJ_EXCAVATION:
    case MP2::OBJ_CAVE:
    case MP2::OBJ_TREEHOUSE:
    case MP2::OBJ_ARCHERHOUSE:
    case MP2::OBJ_GOBLINHUT:
    case MP2::OBJ_DWARFCOTT:
    case MP2::OBJ_HALFLINGHOLE:
    case MP2::OBJ_PEASANTHUT:
    case MP2::OBJ_THATCHEDHUT:
    //
    case MP2::OBJ_MONSTER:
        os << "count           : " << MonsterCount() << std::endl;
        break;

    case MP2::OBJ_HEROES: {
        const Heroes * hero = GetHeroes();
        if ( hero )
            os << hero->String();
    } break;

    case MP2::OBJN_CASTLE:
    case MP2::OBJ_CASTLE: {
        const Castle * castle = world.GetCastle( GetCenter() );
        if ( castle )
            os << castle->String();
    } break;

    default: {
        const MapsIndexes & v = Maps::GetTilesUnderProtection( GetIndex() );
        if ( v.size() ) {
            os << "protection      : ";
            for ( MapsIndexes::const_iterator it = v.begin(); it != v.end(); ++it )
                os << *it << ", ";
            os << std::endl;
        }
        break;
    }
    }

    if ( MP2::isCaptureObject( GetObject( false ) ) ) {
        const CapturedObject & co = world.GetCapturedObject( GetIndex() );

        os << "capture color   : " << Color::String( co.objcol.second ) << std::endl;
        if ( co.guardians.isValid() ) {
            os << "capture guard   : " << co.guardians.GetName() << std::endl << "capture caunt   : " << co.guardians.GetCount() << std::endl;
        }
    }

    os << "----------------:--------" << std::endl;
    return os.str();
}

void Maps::Tiles::FixObject( void )
{
    if ( MP2::OBJ_ZERO == mp2_object ) {
        if ( addons_level1.end() != std::find_if( addons_level1.begin(), addons_level1.end(), TilesAddon::isArtifact ) )
            SetObject( MP2::OBJ_ARTIFACT );
        else if ( addons_level1.end() != std::find_if( addons_level1.begin(), addons_level1.end(), TilesAddon::isResource ) )
            SetObject( MP2::OBJ_RESOURCE );
    }
}

bool Maps::Tiles::GoodForUltimateArtifact( void ) const
{
    return !isWater()
           && ( addons_level1.empty()
                || addons_level1.size() == static_cast<size_t>( std::count_if( addons_level1.begin(), addons_level1.end(), std::ptr_fun( &TilesAddon::isShadow ) ) ) )
           && isPassable( Direction::CENTER, false, true );
}

bool TileIsGround( s32 index, int ground )
{
    return ground == world.GetTiles( index ).GetGround();
}

bool Maps::Tiles::validateWaterRules( bool fromWater ) const
{
    const bool tileIsWater = isWater();
    if ( fromWater )
        return mp2_object == MP2::OBJ_COAST || ( tileIsWater && mp2_object != MP2::OBJ_BOAT );

    // if we're not in water but tile is; allow movement in three cases
    if ( tileIsWater )
        return mp2_object == MP2::OBJ_SHIPWRECK || mp2_object == MP2::OBJ_HEROES || mp2_object == MP2::OBJ_BOAT;

    return true;
}

bool Maps::Tiles::isPassable( int direct, bool fromWater, bool skipfog ) const
{
    if ( !skipfog && isFog( Settings::Get().CurrentColor() ) )
        return false;

    if ( !validateWaterRules( fromWater ) )
        return false;

    return direct & tilePassable;
}

void Maps::Tiles::SetObjectPassable( bool pass )
{
    switch ( GetObject( false ) ) {
    case MP2::OBJ_TROLLBRIDGE:
        if ( pass )
            tilePassable |= Direction::TOP_LEFT;
        else
            tilePassable &= ~Direction::TOP_LEFT;
        break;

    default:
        break;
    }
}

/* check road */
bool Maps::Tiles::isRoad() const
{
    return tileIsRoad || mp2_object == MP2::OBJ_CASTLE;
}

bool Maps::Tiles::isStream( void ) const
{
    return addons_level1.end() != std::find_if( addons_level1.begin(), addons_level1.end(), TilesAddon::isStream );
}

bool Maps::Tiles::isShadow( void ) const
{
    return addons_level1.size() != std::count_if( addons_level1.begin(), addons_level1.end(), TilesAddon::isShadow );
}

bool Maps::Tiles::hasSpriteAnimation() const
{
    return objectTileset & 1;
}

bool Maps::Tiles::isObject( int obj ) const
{
    return obj == mp2_object;
}

uint8_t Maps::Tiles::GetObjectTileset() const
{
    return objectTileset;
}

uint8_t Maps::Tiles::GetObjectSpriteIndex() const
{
    return objectIndex;
}

Maps::TilesAddon * Maps::Tiles::FindFlags( void )
{
    Addons::iterator it = std::find_if( addons_level1.begin(), addons_level1.end(), TilesAddon::isFlag32 );

    if ( it == addons_level1.end() ) {
        it = std::find_if( addons_level2.begin(), addons_level2.end(), TilesAddon::isFlag32 );
        return addons_level2.end() != it ? &( *it ) : NULL;
    }

    return addons_level1.end() != it ? &( *it ) : NULL;
}

/* ICN::FLAGS32 version */
void Maps::Tiles::CaptureFlags32( int obj, int col )
{
    u32 index = 0;

    switch ( col ) {
    case Color::BLUE:
        index = 0;
        break;
    case Color::GREEN:
        index = 1;
        break;
    case Color::RED:
        index = 2;
        break;
    case Color::YELLOW:
        index = 3;
        break;
    case Color::ORANGE:
        index = 4;
        break;
    case Color::PURPLE:
        index = 5;
        break;
    default:
        index = 6;
        break;
    }

    switch ( obj ) {
    case MP2::OBJ_WINDMILL:
        index += 42;
        CorrectFlags32( index, false );
        break;
    case MP2::OBJ_WATERWHEEL:
        index += 14;
        CorrectFlags32( index, false );
        break;
    case MP2::OBJ_MAGICGARDEN:
        index += 42;
        CorrectFlags32( index, false );
        break;

    case MP2::OBJ_MINES:
        index += 14;
        CorrectFlags32( index, true );
        break;
        // case MP2::OBJ_DRAGONCITY:	index += 35; CorrectFlags32(index); break; unused
    case MP2::OBJ_LIGHTHOUSE:
        index += 42;
        CorrectFlags32( index, false );
        break;

    case MP2::OBJ_ALCHEMYLAB: {
        index += 21;
        if ( Maps::isValidDirection( GetIndex(), Direction::TOP ) ) {
            Maps::Tiles & tile = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::TOP ) );
            tile.CorrectFlags32( index, true );
        }
    } break;

    case MP2::OBJ_SAWMILL: {
        index += 28;
        if ( Maps::isValidDirection( GetIndex(), Direction::TOP_RIGHT ) ) {
            Maps::Tiles & tile = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::TOP_RIGHT ) );
            tile.CorrectFlags32( index, true );
        }
    } break;

    case MP2::OBJ_CASTLE: {
        index *= 2;
        if ( Maps::isValidDirection( GetIndex(), Direction::LEFT ) ) {
            Maps::Tiles & tile = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::LEFT ) );
            tile.CorrectFlags32( index, true );
        }

        index += 1;
        if ( Maps::isValidDirection( GetIndex(), Direction::RIGHT ) ) {
            Maps::Tiles & tile = world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::RIGHT ) );
            tile.CorrectFlags32( index, true );
        }
    } break;

    default:
        break;
    }
}

/* correct flags, ICN::FLAGS32 vesion */
void Maps::Tiles::CorrectFlags32( u32 index, bool up )
{
    TilesAddon * taddon = FindFlags();

    // replace flag
    if ( taddon )
        taddon->index = index;
    else if ( up )
        // or new flag
        addons_level2.push_back( TilesAddon( TilesAddon::UPPER, world.GetUniq(), 0x38, index ) );
    else
        // or new flag
        addons_level1.push_back( TilesAddon( TilesAddon::UPPER, world.GetUniq(), 0x38, index ) );
}

void Maps::Tiles::FixedPreload( Tiles & tile )
{
    Addons::iterator it;

    // fix skeleton: left position
    it = std::find_if( tile.addons_level1.begin(), tile.addons_level1.end(), TilesAddon::isSkeletonFix );

    if ( it != tile.addons_level1.end() ) {
        tile.SetObject( MP2::OBJN_SKELETON );
    }

    // fix price loyalty objects.
    if ( Settings::Get().PriceLoyaltyVersion() )
        switch ( tile.GetObject() ) {
        case MP2::OBJ_UNKNW_79:
        case MP2::OBJ_UNKNW_7A:
        case MP2::OBJ_UNKNW_F9:
        case MP2::OBJ_UNKNW_FA: {
            int newobj = MP2::OBJ_ZERO;
            it = std::find_if( tile.addons_level1.begin(), tile.addons_level1.end(), TilesAddon::isX_LOC123 );
            if ( it != tile.addons_level1.end() ) {
                newobj = TilesAddon::GetLoyaltyObject( *it );
            }
            else {
                it = std::find_if( tile.addons_level2.begin(), tile.addons_level2.end(), TilesAddon::isX_LOC123 );
                if ( it != tile.addons_level2.end() )
                    newobj = TilesAddon::GetLoyaltyObject( *it );
            }

            if ( MP2::OBJ_ZERO != newobj )
                tile.SetObject( newobj );
            else {
                DEBUG( DBG_GAME, DBG_WARN, "index: " << tile.GetIndex() );
            }
        } break;

        default:
            break;
        }
}

/* true: if protection or has guardians */
bool Maps::Tiles::CaptureObjectIsProtection( void ) const
{
    const int object = GetObject( false );

    if ( MP2::isCaptureObject( object ) ) {
        if ( MP2::OBJ_CASTLE == object ) {
            Castle * castle = world.GetCastle( GetCenter() );
            if ( castle )
                return castle->GetArmy().isValid();
        }
        else
            return QuantityTroop().isValid();
    }

    return false;
}

void Maps::Tiles::Remove( u32 uniqID )
{
    if ( !addons_level1.empty() )
        addons_level1.Remove( uniqID );
    if ( !addons_level2.empty() )
        addons_level2.Remove( uniqID );

    if ( uniq == uniqID ) {
        objectTileset = 0;
        objectIndex = 255;
        uniq = 0;
    }
}

void Maps::Tiles::UpdateObjectSprite( uint32_t uniqID, uint8_t rawTileset, uint8_t newTileset, uint8_t newIndex, bool replace )
{
    for ( Addons::iterator it = addons_level1.begin(); it != addons_level1.end(); ++it ) {
        if ( it->uniq == uniqID && ( it->object >> 2 ) == rawTileset ) {
            it->object = replace ? newTileset : it->object + newTileset;
            it->index = replace ? newIndex : it->index + newIndex;
        }
    }
    for ( Addons::iterator it2 = addons_level2.begin(); it2 != addons_level2.end(); ++it2 ) {
        if ( it2->uniq == uniqID && ( it2->object >> 2 ) == rawTileset ) {
            it2->object = replace ? newTileset : it2->object + newTileset;
            it2->index = replace ? newIndex : it2->index + newIndex;
        }
    }

    if ( uniq == uniqID && ( objectTileset >> 2 ) == rawTileset ) {
        objectTileset = replace ? newTileset : objectTileset + newTileset;
        objectIndex = replace ? newIndex : objectIndex + newIndex;
    }
}

void Maps::Tiles::RemoveObjectSprite( void )
{
    switch ( GetObject() ) {
    case MP2::OBJ_JAIL:
        RemoveJailSprite();
        tilePassable = DIRECTION_ALL;
        break;
    case MP2::OBJ_BARRIER:
        RemoveBarrierSprite();
        tilePassable = DIRECTION_ALL;
        break;

    default:
        // remove shadow sprite from left cell
        if ( Maps::isValidDirection( GetIndex(), Direction::LEFT ) )
            world.GetTiles( Maps::GetDirectionIndex( GetIndex(), Direction::LEFT ) ).Remove( uniq );

        Remove( uniq );
        break;
    }
}

void Maps::Tiles::RemoveBarrierSprite( void )
{
    // remove left sprite
    if ( Maps::isValidDirection( GetIndex(), Direction::LEFT ) ) {
        const s32 left = Maps::GetDirectionIndex( GetIndex(), Direction::LEFT );
        world.GetTiles( left ).Remove( uniq );
    }

    Remove( uniq );
}

void Maps::Tiles::RemoveJailSprite( void )
{
    // remove left sprite
    if ( Maps::isValidDirection( GetIndex(), Direction::LEFT ) ) {
        const s32 left = Maps::GetDirectionIndex( GetIndex(), Direction::LEFT );
        world.GetTiles( left ).Remove( uniq );

        // remove left left sprite
        if ( Maps::isValidDirection( left, Direction::LEFT ) )
            world.GetTiles( Maps::GetDirectionIndex( left, Direction::LEFT ) ).Remove( uniq );
    }

    // remove top sprite
    if ( Maps::isValidDirection( GetIndex(), Direction::TOP ) ) {
        const s32 top = Maps::GetDirectionIndex( GetIndex(), Direction::TOP );
        Maps::Tiles & topTile = world.GetTiles( top );
        topTile.Remove( uniq );
        topTile.SetObject( MP2::OBJ_ZERO );
        topTile.FixObject();

        // remove top left sprite
        if ( Maps::isValidDirection( top, Direction::LEFT ) ) {
            Maps::Tiles & leftTile = world.GetTiles( Maps::GetDirectionIndex( top, Direction::LEFT ) );
            leftTile.Remove( uniq );
            leftTile.SetObject( MP2::OBJ_ZERO );
            leftTile.FixObject();
        }
    }

    Remove( uniq );
}

void Maps::Tiles::UpdateAbandoneMineSprite( Tiles & tile )
{
    if ( tile.uniq ) {
        const int type = tile.QuantityResourceCount().first;

        Tiles::UpdateAbandoneMineLeftSprite( tile.objectTileset, tile.objectIndex, type );
        for ( Addons::iterator it = tile.addons_level1.begin(); it != tile.addons_level1.end(); ++it )
            Tiles::UpdateAbandoneMineLeftSprite( it->object, it->index, type );

        if ( Maps::isValidDirection( tile.GetIndex(), Direction::RIGHT ) ) {
            Tiles & tile2 = world.GetTiles( Maps::GetDirectionIndex( tile.GetIndex(), Direction::RIGHT ) );
            TilesAddon * mines = tile2.FindAddonLevel1( tile.uniq );

            if ( mines )
                Tiles::UpdateAbandoneMineRightSprite( mines->object, mines->index );

            if ( tile2.GetObject() == MP2::OBJN_ABANDONEDMINE ) {
                tile2.SetObject( MP2::OBJN_MINES );
                Tiles::UpdateAbandoneMineRightSprite( tile2.objectTileset, tile2.objectIndex );
            }
        }
    }

    if ( Maps::isValidDirection( tile.GetIndex(), Direction::LEFT ) ) {
        Tiles & tile2 = world.GetTiles( Maps::GetDirectionIndex( tile.GetIndex(), Direction::LEFT ) );
        if ( tile2.GetObject() == MP2::OBJN_ABANDONEDMINE )
            tile2.SetObject( MP2::OBJN_MINES );
    }

    if ( Maps::isValidDirection( tile.GetIndex(), Direction::TOP ) ) {
        Tiles & tile2 = world.GetTiles( Maps::GetDirectionIndex( tile.GetIndex(), Direction::TOP ) );
        if ( tile2.GetObject() == MP2::OBJN_ABANDONEDMINE )
            tile2.SetObject( MP2::OBJN_MINES );

        if ( Maps::isValidDirection( tile2.GetIndex(), Direction::LEFT ) ) {
            Tiles & tile3 = world.GetTiles( Maps::GetDirectionIndex( tile2.GetIndex(), Direction::LEFT ) );
            if ( tile3.GetObject() == MP2::OBJN_ABANDONEDMINE )
                tile3.SetObject( MP2::OBJN_MINES );
        }

        if ( Maps::isValidDirection( tile2.GetIndex(), Direction::RIGHT ) ) {
            Tiles & tile3 = world.GetTiles( Maps::GetDirectionIndex( tile2.GetIndex(), Direction::RIGHT ) );
            if ( tile3.GetObject() == MP2::OBJN_ABANDONEDMINE )
                tile3.SetObject( MP2::OBJN_MINES );
        }
    }
}

void Maps::Tiles::UpdateRNDArtifactSprite( Tiles & tile )
{
    Artifact art;

    switch ( tile.GetObject() ) {
    case MP2::OBJ_RNDARTIFACT:
        art = Artifact::Rand( Artifact::ART_LEVEL123 );
        break;
    case MP2::OBJ_RNDARTIFACT1:
        art = Artifact::Rand( Artifact::ART_LEVEL1 );
        break;
    case MP2::OBJ_RNDARTIFACT2:
        art = Artifact::Rand( Artifact::ART_LEVEL2 );
        break;
    case MP2::OBJ_RNDARTIFACT3:
        art = Artifact::Rand( Artifact::ART_LEVEL3 );
        break;
    default:
        return;
    }

    if ( !art.isValid() ) {
        DEBUG( DBG_GAME, DBG_WARN, "unknown artifact" );
        return;
    }

    const uint8_t index = art.IndexSprite();

    tile.SetObject( MP2::OBJ_ARTIFACT );
    tile.objectIndex = index;

    // replace artifact shadow
    if ( Maps::isValidDirection( tile.GetIndex(), Direction::LEFT ) ) {
        Maps::Tiles & left_tile = world.GetTiles( Maps::GetDirectionIndex( tile.GetIndex(), Direction::LEFT ) );
        Maps::TilesAddon * shadow = left_tile.FindAddonLevel1( tile.uniq );

        if ( shadow )
            shadow->index = index - 1;
    }
}

void Maps::Tiles::UpdateRNDResourceSprite( Tiles & tile )
{
    tile.SetObject( MP2::OBJ_RESOURCE );
    tile.objectIndex = Resource::GetIndexSprite( Resource::Rand() );

    // replace shadow artifact
    if ( Maps::isValidDirection( tile.GetIndex(), Direction::LEFT ) ) {
        Maps::Tiles & left_tile = world.GetTiles( Maps::GetDirectionIndex( tile.GetIndex(), Direction::LEFT ) );
        Maps::TilesAddon * shadow = left_tile.FindAddonLevel1( tile.uniq );

        if ( shadow )
            shadow->index = tile.objectIndex - 1;
    }
}

std::pair<int, int> Maps::Tiles::GetMonsterSpriteIndices( const Tiles & tile, uint32_t monsterIndex )
{
    const int tileIndex = tile.GetIndex();
    int attackerIndex = -1;

    // scan hero around
    const MapsIndexes & v = ScanAroundObject( tileIndex, MP2::OBJ_HEROES );
    for ( MapsIndexes::const_iterator it = v.begin(); it != v.end(); ++it ) {
        const Tiles & heroTile = world.GetTiles( *it );
        const Heroes * hero = heroTile.GetHeroes();
        if ( hero == NULL ) { // no a hero? How can it be?!
            continue;
        }

        if ( tile.isWater() == heroTile.isWater() ) {
            attackerIndex = *it;
            break;
        }
    }

    std::pair<int, int> spriteIndices( monsterIndex * 9, -1 );

    // draw attack sprite
    if ( attackerIndex != -1 && !Settings::Get().ExtWorldOnlyFirstMonsterAttack() ) {
        spriteIndices.first += 7;

        switch ( Maps::GetDirection( tileIndex, attackerIndex ) ) {
        case Direction::TOP_LEFT:
        case Direction::LEFT:
        case Direction::BOTTOM_LEFT:
            spriteIndices.first += 1;
            break;
        default:
            break;
        }
    }
    else {
        const Point mp = Maps::GetPoint( tileIndex );
        spriteIndices.second = monsterIndex * 9 + 1 + monsterAnimationSequence[( Game::MapsAnimationFrame() + mp.x * mp.y ) % ARRAY_COUNT( monsterAnimationSequence )];
    }
    return spriteIndices;
}

bool Maps::Tiles::isFog( int colors ) const
{
    // colors may be the union friends
    return ( fog_colors & colors ) == colors;
}

void Maps::Tiles::ClearFog( int colors )
{
    fog_colors &= ~colors;
}

void Maps::Tiles::RedrawFogs( fheroes2::Image & dst, int color ) const
{
    const Interface::GameArea & area = Interface::Basic::Get().GetGameArea();
    const Point mp = Maps::GetPoint( GetIndex() );

    // get direction around foga
    int around = 0;
    const Directions directions = Direction::All();

    for ( Directions::const_iterator it = directions.begin(); it != directions.end(); ++it )
        if ( !Maps::isValidDirection( GetIndex(), *it ) || world.GetTiles( Maps::GetDirectionIndex( GetIndex(), *it ) ).isFog( color ) )
            around |= *it;

    if ( isFog( color ) )
        around |= Direction::CENTER;

    // TIL::CLOF32
    if ( DIRECTION_ALL == around ) {
        const fheroes2::Image & sf = fheroes2::AGG::GetTIL( TIL::CLOF32, GetIndex() % 4, 0 );
        area.BlitOnTile( dst, sf, 0, 0, mp );
    }
    else {
        u32 index = 0;
        bool revert = false;

        // see ICN::CLOP32: sprite 10
        if ( ( around & Direction::CENTER ) && !( around & ( Direction::TOP | Direction::BOTTOM | Direction::LEFT | Direction::RIGHT ) ) ) {
            index = 10;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 6, 7, 8
            if ( ( contains( around, Direction::CENTER | Direction::TOP ) ) && !( around & ( Direction::BOTTOM | Direction::LEFT | Direction::RIGHT ) ) ) {
            index = 6;
            revert = false;
        }
        else if ( ( contains( around, Direction::CENTER | Direction::RIGHT ) ) && !( around & ( Direction::TOP | Direction::BOTTOM | Direction::LEFT ) ) ) {
            index = 7;
            revert = false;
        }
        else if ( ( contains( around, Direction::CENTER | Direction::LEFT ) ) && !( around & ( Direction::TOP | Direction::BOTTOM | Direction::RIGHT ) ) ) {
            index = 7;
            revert = true;
        }
        else if ( ( contains( around, Direction::CENTER | Direction::BOTTOM ) ) && !( around & ( Direction::TOP | Direction::LEFT | Direction::RIGHT ) ) ) {
            index = 8;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 9, 29
            if ( ( contains( around, DIRECTION_CENTER_COL ) ) && !( around & ( Direction::LEFT | Direction::RIGHT ) ) ) {
            index = 9;
            revert = false;
        }
        else if ( ( contains( around, DIRECTION_CENTER_ROW ) ) && !( around & ( Direction::TOP | Direction::BOTTOM ) ) ) {
            index = 29;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 15, 22
            if ( around == ( DIRECTION_ALL & ( ~Direction::TOP_RIGHT ) ) ) {
            index = 15;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~Direction::TOP_LEFT ) ) ) {
            index = 15;
            revert = true;
        }
        else if ( around == ( DIRECTION_ALL & ( ~Direction::BOTTOM_RIGHT ) ) ) {
            index = 22;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~Direction::BOTTOM_LEFT ) ) ) {
            index = 22;
            revert = true;
        }
        else
            // see ICN::CLOP32: sprite 16, 17, 18, 23
            if ( around == ( DIRECTION_ALL & ( ~( Direction::TOP_RIGHT | Direction::BOTTOM_RIGHT ) ) ) ) {
            index = 16;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~( Direction::TOP_LEFT | Direction::BOTTOM_LEFT ) ) ) ) {
            index = 16;
            revert = true;
        }
        else if ( around == ( DIRECTION_ALL & ( ~( Direction::TOP_RIGHT | Direction::BOTTOM_LEFT ) ) ) ) {
            index = 17;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~( Direction::TOP_LEFT | Direction::BOTTOM_RIGHT ) ) ) ) {
            index = 17;
            revert = true;
        }
        else if ( around == ( DIRECTION_ALL & ( ~( Direction::TOP_LEFT | Direction::TOP_RIGHT ) ) ) ) {
            index = 18;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~( Direction::BOTTOM_LEFT | Direction::BOTTOM_RIGHT ) ) ) ) {
            index = 23;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 13, 14
            if ( around == ( DIRECTION_ALL & ( ~DIRECTION_TOP_RIGHT_CORNER ) ) ) {
            index = 13;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~DIRECTION_TOP_LEFT_CORNER ) ) ) {
            index = 13;
            revert = true;
        }
        else if ( around == ( DIRECTION_ALL & ( ~DIRECTION_BOTTOM_RIGHT_CORNER ) ) ) {
            index = 14;
            revert = false;
        }
        else if ( around == ( DIRECTION_ALL & ( ~DIRECTION_BOTTOM_LEFT_CORNER ) ) ) {
            index = 14;
            revert = true;
        }
        else
            // see ICN::CLOP32: sprite 11, 12
            if ( contains( around, Direction::CENTER | Direction::LEFT | Direction::BOTTOM_LEFT | Direction::BOTTOM )
                 && !( around & ( Direction::TOP | Direction::RIGHT ) ) ) {
            index = 11;
            revert = false;
        }
        else if ( contains( around, Direction::CENTER | Direction::RIGHT | Direction::BOTTOM_RIGHT | Direction::BOTTOM )
                  && !( around & ( Direction::TOP | Direction::LEFT ) ) ) {
            index = 11;
            revert = true;
        }
        else if ( contains( around, Direction::CENTER | Direction::LEFT | Direction::TOP_LEFT | Direction::TOP )
                  && !( around & ( Direction::BOTTOM | Direction::RIGHT ) ) ) {
            index = 12;
            revert = false;
        }
        else if ( contains( around, Direction::CENTER | Direction::RIGHT | Direction::TOP_RIGHT | Direction::TOP )
                  && !( around & ( Direction::BOTTOM | Direction::LEFT ) ) ) {
            index = 12;
            revert = true;
        }
        else
            // see ICN::CLOP32: sprite 19, 20, 22
            if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::TOP | Direction::TOP_LEFT )
                 && !( around & ( Direction::BOTTOM_LEFT | Direction::BOTTOM_RIGHT | Direction::TOP_RIGHT ) ) ) {
            index = 19;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::TOP | Direction::TOP_RIGHT )
                  && !( around & ( Direction::BOTTOM_LEFT | Direction::BOTTOM_RIGHT | Direction::TOP_LEFT ) ) ) {
            index = 19;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::TOP | Direction::BOTTOM_LEFT )
                  && !( around & ( Direction::TOP_RIGHT | Direction::BOTTOM_RIGHT | Direction::TOP_LEFT ) ) ) {
            index = 20;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::TOP | Direction::BOTTOM_RIGHT )
                  && !( around & ( Direction::TOP_RIGHT | Direction::BOTTOM_LEFT | Direction::TOP_LEFT ) ) ) {
            index = 20;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::TOP )
                  && !( around & ( Direction::TOP_RIGHT | Direction::BOTTOM_RIGHT | Direction::BOTTOM_LEFT | Direction::TOP_LEFT ) ) ) {
            index = 22;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 24, 25, 26, 30
            if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::BOTTOM_LEFT ) && !( around & ( Direction::TOP | Direction::BOTTOM_RIGHT ) ) ) {
            index = 24;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM | Direction::BOTTOM_RIGHT ) && !( around & ( Direction::TOP | Direction::BOTTOM_LEFT ) ) ) {
            index = 24;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::LEFT | Direction::TOP_LEFT ) && !( around & ( Direction::RIGHT | Direction::BOTTOM_LEFT ) ) ) {
            index = 25;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::RIGHT | Direction::TOP_RIGHT ) && !( around & ( Direction::LEFT | Direction::BOTTOM_RIGHT ) ) ) {
            index = 25;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::BOTTOM_LEFT | Direction::LEFT ) && !( around & ( Direction::RIGHT | Direction::TOP_LEFT ) ) ) {
            index = 26;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::BOTTOM_RIGHT | Direction::RIGHT ) && !( around & ( Direction::LEFT | Direction::TOP_RIGHT ) ) ) {
            index = 26;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::TOP_LEFT | Direction::TOP ) && !( around & ( Direction::BOTTOM | Direction::TOP_RIGHT ) ) ) {
            index = 30;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::TOP_RIGHT | Direction::TOP ) && !( around & ( Direction::BOTTOM | Direction::TOP_LEFT ) ) ) {
            index = 30;
            revert = true;
        }
        else
            // see ICN::CLOP32: sprite 27, 28
            if ( contains( around, Direction::CENTER | Direction::BOTTOM | Direction::LEFT )
                 && !( around & ( Direction::TOP | Direction::TOP_RIGHT | Direction::RIGHT | Direction::BOTTOM_LEFT ) ) ) {
            index = 27;
            revert = false;
        }
        else if ( contains( around, Direction::CENTER | Direction::BOTTOM | Direction::RIGHT )
                  && !( around & ( Direction::TOP | Direction::TOP_LEFT | Direction::LEFT | Direction::BOTTOM_RIGHT ) ) ) {
            index = 27;
            revert = true;
        }
        else if ( contains( around, Direction::CENTER | Direction::LEFT | Direction::TOP )
                  && !( around & ( Direction::TOP_LEFT | Direction::RIGHT | Direction::BOTTOM | Direction::BOTTOM_RIGHT ) ) ) {
            index = 28;
            revert = false;
        }
        else if ( contains( around, Direction::CENTER | Direction::RIGHT | Direction::TOP )
                  && !( around & ( Direction::TOP_RIGHT | Direction::LEFT | Direction::BOTTOM | Direction::BOTTOM_LEFT ) ) ) {
            index = 28;
            revert = true;
        }
        else
            // see ICN::CLOP32: sprite 31, 32, 33
            if ( contains( around, DIRECTION_CENTER_ROW | Direction::TOP ) && !( around & ( Direction::BOTTOM | Direction::TOP_LEFT | Direction::TOP_RIGHT ) ) ) {
            index = 31;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::RIGHT ) && !( around & ( Direction::LEFT | Direction::TOP_RIGHT | Direction::BOTTOM_RIGHT ) ) ) {
            index = 32;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | Direction::LEFT ) && !( around & ( Direction::RIGHT | Direction::TOP_LEFT | Direction::BOTTOM_LEFT ) ) ) {
            index = 32;
            revert = true;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | Direction::BOTTOM ) && !( around & ( Direction::TOP | Direction::BOTTOM_LEFT | Direction::BOTTOM_RIGHT ) ) ) {
            index = 33;
            revert = false;
        }
        else
            // see ICN::CLOP32: sprite 0, 1, 2, 3, 4, 5
            if ( contains( around, DIRECTION_CENTER_ROW | DIRECTION_BOTTOM_ROW ) && !( around & ( Direction::TOP ) ) ) {
            index = ( GetIndex() % 2 ) ? 0 : 1;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_ROW | DIRECTION_TOP_ROW ) && !( around & ( Direction::BOTTOM ) ) ) {
            index = ( GetIndex() % 2 ) ? 4 : 5;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | DIRECTION_LEFT_COL ) && !( around & ( Direction::RIGHT ) ) ) {
            index = ( GetIndex() % 2 ) ? 2 : 3;
            revert = false;
        }
        else if ( contains( around, DIRECTION_CENTER_COL | DIRECTION_RIGHT_COL ) && !( around & ( Direction::LEFT ) ) ) {
            index = ( GetIndex() % 2 ) ? 2 : 3;
            revert = true;
        }
        // unknown
        else {
            DEBUG( DBG_GAME, DBG_WARN, "Invalid direction for fog: " << around );
            const fheroes2::Image & sf = fheroes2::AGG::GetTIL( TIL::CLOF32, GetIndex() % 4, 0 );
            area.BlitOnTile( dst, sf, 0, 0, mp );
            return;
        }

        const fheroes2::Sprite & sprite = fheroes2::AGG::GetICN( ICN::CLOP32, index );
        area.BlitOnTile( dst, sprite, ( revert ? sprite.x() + TILEWIDTH - sprite.width() : sprite.x() ), sprite.y(), mp, revert );
    }
}

StreamBase & Maps::operator<<( StreamBase & msg, const TilesAddon & ta )
{
    return msg << ta.level << ta.uniq << ta.object << ta.index << ta.tmp;
}

StreamBase & Maps::operator>>( StreamBase & msg, TilesAddon & ta )
{
    msg >> ta.level >> ta.uniq >> ta.object >> ta.index >> ta.tmp;
    if ( FORMAT_VERSION_080_RELEASE > Game::GetLoadVersion() ) {
        switch ( ta.object ) {
        case 0x11:
            ta.object = 0xA4;
            ta.index = 116;
            break;
        case 0x12:
            ta.object = 0xA4;
            ta.index = 119;
            break;
        case 0x13:
            ta.object = 0xA4;
            ta.index = 122;
            break;
        case 0x14:
            ta.object = 0xA4;
            ta.index = 15;
            break;
        case 0x15:
            ta.object = 0xB8;
            ta.index = 19;
            break;
        default:
            break;
        }
    }
    return msg;
}

StreamBase & Maps::operator<<( StreamBase & msg, const Tiles & tile )
{
    return msg << tile.maps_index << tile.pack_sprite_index << tile.tilePassable << tile.uniq << tile.objectTileset << tile.objectIndex << tile.mp2_object
               << tile.fog_colors << tile.quantity1 << tile.quantity2 << tile.quantity3 << tile.heroID << tile.tileIsRoad << tile.addons_level1 << tile.addons_level2;
}

StreamBase & Maps::operator>>( StreamBase & msg, Tiles & tile )
{
    msg >> tile.maps_index >> tile.pack_sprite_index >> tile.tilePassable;

    // Backwards compatibility with saves pre-0.8.2 release
    if ( FORMAT_VERSION_082_RELEASE > Game::GetLoadVersion() ) {
        tile.uniq = 0;
        tile.objectTileset = 0;
        tile.objectIndex = 255;
        tile.tileIsRoad = 0;
        msg >> tile.mp2_object >> tile.fog_colors >> tile.quantity1 >> tile.quantity2;

        tile.heroID = 0;
        tile.quantity3 = 0;
        if ( tile.mp2_object == MP2::OBJ_HEROES ) {
            msg >> tile.heroID;
        }
        else {
            msg >> tile.quantity3;
        }

        msg >> tile.addons_level1 >> tile.addons_level2;
        for ( const Maps::TilesAddon & addon : tile.addons_level1 ) {
            if ( addon.isRoad() ) {
                tile.tileIsRoad = true;
                break;
            }
        }

        tile.AddonsSort();
    }
    else {
        msg >> tile.uniq >> tile.objectTileset >> tile.objectIndex >> tile.mp2_object >> tile.fog_colors >> tile.quantity1 >> tile.quantity2 >> tile.quantity3
            >> tile.heroID >> tile.tileIsRoad >> tile.addons_level1 >> tile.addons_level2;
    }

    return msg;
}