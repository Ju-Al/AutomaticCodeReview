/* Copyright (C) 2009-2012, Stefan Hacker <dD0t@users.sourceforge.net>
static int fetch(float *avatar_pos, float *avatar_front, float *avatar_top, float *camera_pos, float *camera_front, float *camera_top, std::string &, std::wstring &) {
	char state;

	/*
		This plugin uses the following Variables:

			Address			Type	Description
			===================================
			0x013F79CC		float	Y-Coordinate
			0x013E8CF4		float	X-Coordinate
			0x013E8CF8		float	Z-Coordinate
			0x013F9E20		float	Horizontal view
			0x013F9E1C		float	Vertical view
			0x013E8D18		byte	Magic value (32 ingame / 0 spectating)
	*/
	ok = peekProc((BYTE *) 0x013E8D18, &state, 1);
	if (! ok)
	if (state != 32)
		return true; // If this magic value is !=32 we are spectating, so switch off PA
	ok = peekProc((BYTE *) 0x013F79CC, avatar_pos+1, 4) &&	//Y
	     peekProc((BYTE *) 0x013E8CF4, avatar_pos, 4) &&	//X
	     peekProc((BYTE *) 0x013E8CF8, avatar_pos+2, 4) && //Z
	     peekProc((BYTE *) 0x013F9E20, &viewHor, 4) && //Hor
	     peekProc((BYTE *) 0x013F9E1C, &viewVer, 4); //Ver
   Copyright (C) 2005-2012, Thorvald Natvig <thorvald@natvig.com>
   Copyright (C) 2012, Jan Klass <kissaki@kcode.de>

   All rights reserved.

   Redistribution and use in source and binary forms, with or without
   modification, are permitted provided that the following conditions
   are met:

   - Redistributions of source code must retain the above copyright notice,
     this list of conditions and the following disclaimer.
   - Redistributions in binary form must reproduce the above copyright notice,
     this list of conditions and the following disclaimer in the documentation
     and/or other materials provided with the distribution.
   - Neither the name of the Mumble Developers nor the names of its
     contributors may be used to endorse or promote products derived from this
     software without specific prior written permission.

   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
   ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
   LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
   A PARTICULAR PURPOSE ARE DISCLAIMED.  IN NO EVENT SHALL THE FOUNDATION OR
   CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
   EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
   PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
   PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
   LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
   NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
   SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/*
    This plugin scans the following address locations:

        Address         Type      Description
        ===================================
        0x013E8CF4      float     X-Coordinate
        0x013E8CF8      float     Y-Coordinate
        0x013E8CFC      float     Z-Coordinate
        // The XYZ coordinates repeat at 0x013F79C8, 0x013F79CC, 0x013F79D0
        // and seem to always hold the same values. It is not the camera vs
        // character, as with third person they still hold the same values.
        0x013F9E20      float     Horizontal view in degree
        0x013F9E1C      float     Vertical view in degree
        0x013E8DFC      byte      Team; 0: not in-game (in menu etc), 1: allies, 2: axis, 3: spectator
        0x009FFD30      char[40]  last connected to host IP
        0x010B4908      char[32]  map shortname only

        As reference, as help for future manual mem scanning:
        0x010B48C8      char[32]  map fullpath (relative to gamedir)
        0x03A54C9C      char[16]  last download host IP (wwwdownload)
        There seem to be a ton of memory addresses valid for the team value.

        For position coordinates:
        X-Value is increasing when heading east
                   decreasing when heading west
        Y-Value is increasing when heading north
                   decreasing when heading south
        Z-Value is increasing when going up
                   decreasing when going down
        1 unit = 1 meter (not confirmed)

        For the view angle values:
        Vertical view 0 when centered
                    -90 when looking up
                    +90 when looking down
        Increasing when looking down.

        Horizontal is 90 when facing north
                       0 when facing east
                     -90 when facing south
                     +/-180 when facing west
        Increasing when turning left.
*/

#include "../mumble_plugin_win32.h"

static int fetch(float *avatar_pos, float *avatar_front, float *avatar_top, float *camera_pos, float *camera_front, float *camera_top, std::string &context, std::wstring &) {
	float viewHor, viewVer;
	char team;

	for (int i=0;i<3;i++)
		avatar_pos[i] = avatar_front[i] = avatar_top[i] = camera_pos[i] = camera_front[i] = camera_top[i] = 0.0f;

	bool ok;
	ok = peekProc((BYTE *) 0x013E8DFC, &team, 1);
	if (!ok)
		return false;

	// We dont use pos-audio in menus
	if (team == 0 || team == 3)
		return true;

	ok = peekProc((BYTE *) 0x013E8CF4, avatar_pos, 4) &&   // X
		peekProc((BYTE *) 0x013E8CF8, avatar_pos+1, 4) &&  // Y
		peekProc((BYTE *) 0x013E8CFC, avatar_pos+2, 4) &&  // Z
		peekProc((BYTE *) 0x013F9E20, &viewHor, 4) &&      // Hor-Angle
		peekProc((BYTE *) 0x013F9E1C, &viewVer, 4);        // Ver-Angle

	if (!ok)
		return false;

	avatar_top[2] = -1; // Head movement is in front vector

	// Calculate view unit vector
	viewVer *= static_cast<float>(M_PI / 180.0f);
	viewHor *= static_cast<float>(M_PI / 180.0f);

	avatar_front[0] = cos(viewVer) * cos(viewHor);
	avatar_front[1] = -sin(viewVer);
	avatar_front[2] = cos(viewVer) * sin(viewHor);

	for (int i=0;i<3;i++) {
		camera_pos[i] = avatar_pos[i];
		camera_front[i] = avatar_front[i];
		camera_top[i] = avatar_top[i];
	}

	// Context - concatenated server-ip mapname and team value
	char hostip[32];
	char mapname[40];
	ok = peekProc((BYTE *) 0x009FFD30, hostip, sizeof(hostip)) &&
		peekProc((BYTE *) 0x010B4908, mapname, sizeof(hostip));
	hostip[sizeof(hostip)] = '\0';
	mapname[sizeof(mapname)] = '\0';
	context = std::string(hostip) + mapname + (char)(team + 0x30);

	return true;
}

static int trylock(const std::multimap<std::wstring, unsigned long long int> &pids) {
	if (! initialize(pids, L"ET.exe"))
		return false;

	float apos[3], afront[3], atop[3], cpos[3], cfront[3], ctop[3];
	std::string context;
	std::wstring identity;

	if (fetch(apos, afront, atop, cpos, cfront, ctop, context, identity)) {
		return true;
	} else {
		generic_unlock();
		return false;
	}
}

static const std::wstring longdesc() {
	return std::wstring(L"Supports Wolfenstien: Enemy Territory v2.60b. No context or identity support yet.");
}

static std::wstring description(L"Wolfenstein: Enemy Territory v2.60b");
static std::wstring shortname(L"Wolfenstein: Enemy Territory");

static int trylock1() {
	return trylock(std::multimap<std::wstring, unsigned long long int>());
}

static MumblePlugin wolfetplug = {
	MUMBLE_PLUGIN_MAGIC,
	description,
	shortname,
	NULL,
	NULL,
	trylock1,
	generic_unlock,
	longdesc,
	fetch
};

static MumblePlugin2 wolfetplug2 = {
	MUMBLE_PLUGIN_MAGIC_2,
	MUMBLE_PLUGIN_VERSION,
	trylock
};

extern "C" __declspec(dllexport) MumblePlugin *getMumblePlugin() {
	return &wolfetplug;
}

extern "C" __declspec(dllexport) MumblePlugin2 *getMumblePlugin2() {
	return &wolfetplug2;
}
