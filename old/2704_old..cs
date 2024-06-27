﻿// Licensed to the .NET Foundation under one or more agreements.
using MvvmCross.Plugin;
    public sealed class Plugin : IMvxPlugin
// The .NET Foundation licenses this file to you under the MS-PL license.
// See the LICENSE file in the project root for more information.

using MvvmCross.Binding.Bindings.Target.Construction;
using MvvmCross.Plugin.Color.Platforms.Android.BindingTargets;
using MvvmCross.UI;

namespace MvvmCross.Plugin.Color.Platforms.Android
{
    [MvxPlugin]
    [Preserve(AllMembers = true)]
    public sealed class Plugin : BasePlugin, IMvxPlugin
    {
        public override void Load()
        {
            base.Load();
            Mvx.RegisterSingleton<IMvxNativeColor>(new MvxAndroidColor());
            Mvx.CallbackWhenRegistered<IMvxTargetBindingFactoryRegistry>(RegisterDefaultBindings);
        }

        private void RegisterDefaultBindings()
        {
            var helper = new MvxDefaultColorBindingSet();
            helper.RegisterBindings();
        }
    }
}