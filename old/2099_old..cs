            {
                MvxTrace.Trace("No view model type finder available - assuming we are looking for a splash screen - returning null");
                return typeof(MvxNullViewModel);
            }

            return associatedTypeFinder.FindTypeOrNull(activityType);
        }

        public static bool IsFragmentCacheable(this Type fragmentType, Type fragmentActivityParentType)
        {
            if (!fragmentType.HasBasePresentationAttribute())
                return false;

            var attribute = fragmentType.GetBasePresentationAttribute();
            if (attribute is MvxFragmentPresentationAttribute fragmentAttribute)
                return fragmentAttribute.IsCacheableFragment;
            else
                return false;
        }
    }
}