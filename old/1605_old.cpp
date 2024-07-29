        {
            msg_info("TypeInfoRegistry") << " Promoting typeinfo "<< id << " from " << typeinfos[id]->name() << " to " << info->name();
            info->setCompilationTarget(compilationTarget);
            typeinfos[id] = info;
            return 2;
        }
        return -1;
    }
    if( info->ValidInfo() )
    {
        msg_info("TypeInfoRegistry") << " Registering a complete type info at "  << id << " => " << info->name();
    }
    else
    {
        msg_warning("TypeInfoRegistry") << " Registering a partial new type info at "  << id << " => " << info->name();
    }
    info->setCompilationTarget(compilationTarget);
    typeinfos[id] = info;
    return 1;
}


}
