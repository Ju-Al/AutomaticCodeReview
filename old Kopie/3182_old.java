/*
 * Copyright (C) 2015-2016 Federico Tomassetti
 * Copyright (C) 2017-2020 The JavaParser Team.
 *
 * This file is part of JavaParser.
 *
 * JavaParser can be used either under the terms of
 * a) the GNU Lesser General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 * b) the terms of the Apache License
 *
 * You should have received a copy of both licenses in LICENCE.LGPL and
 * LICENCE.APACHE. Please refer to those files for details.
 *
 * JavaParser is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 */

package com.github.javaparser.symbolsolver.model.typesystem;

import com.github.javaparser.resolution.declarations.ResolvedTypeParameterDeclaration;
import com.github.javaparser.resolution.types.ResolvedArrayType;
import com.github.javaparser.resolution.types.ResolvedPrimitiveType;
import com.github.javaparser.resolution.types.ResolvedReferenceType;
import com.github.javaparser.resolution.types.ResolvedType;
import com.github.javaparser.resolution.types.ResolvedTypeVariable;
import com.github.javaparser.resolution.types.ResolvedVoidType;
import com.github.javaparser.resolution.types.ResolvedWildcard;
import com.google.common.base.Preconditions;

import java.util.Map;
import java.util.Objects;
import java.util.function.Function;
import java.util.function.Supplier;

public class LazyType implements ResolvedType {

    private final Supplier<ResolvedType> supplier;

    private ResolvedType concrete;

    public LazyType(Function<Void, ResolvedType> provider) {
        Preconditions.checkNotNull(provider, "The provider can not be null!");
        this.supplier = () -> provider.apply(null);
    }

    public LazyType(Supplier<ResolvedType> supplier) {
        Preconditions.checkNotNull(supplier, "The supplier can not be null!");
        this.supplier = supplier;
    }

        return concrete;
    }
    /**
     * Get the concrete {@link ResolvedType} for the current {@link LazyType}.
     *
     * @return The {@link ResolvedType} if already resolved, otherwise empty.
     */
    public Optional<ResolvedType> getConcreteType() {
        return Optional.ofNullable(concrete);
    private ResolvedType getType() {
     * Get the type wrapped by a {@link LazyType}.
     *
     * This method is used to get and unpack the {@link LazyType} when they are nested.
     *
     * {@code
     *      LazyType typeA = new LazyType(() -> ResolvedVoidType.INSTANCE);
     *      LazyType typeB = new LazyType(() -> typeA);
     *      LazyType typeC = new LazyType(() -> typeB);
     * }
     *
     * In this examples, by calling {@code getType()} we would expected all of them to return {@link ResolvedVoidType}
     *
     * @return The concrete type for the {@link LazyType}.
     */
    public ResolvedType getType() {
        if (concrete == null) {
            concrete = supplier.get();
        }

        // If the concrete value is a lazy type, then get the inner type.
        if (concrete instanceof LazyType) {
            return ((LazyType) concrete).getType();
        } else {
            return concrete;
        }
    }

    @Override
    public boolean isArray() {
        return getType().isArray();
    }

    @Override
    public int arrayLevel() {
        return getType().arrayLevel();
    }

    @Override
    public boolean isPrimitive() {
        return getType().isPrimitive();
    }

    @Override
    public boolean isNull() {
        return getType().isNull();
    }

    @Override
    public boolean isReference() {
        return getType().isReference();
    }

    @Override
    public boolean isReferenceType() {
        return getType().isReferenceType();
    }

    @Override
    public boolean isVoid() {
        return getType().isVoid();
    }

    @Override
    public boolean isTypeVariable() {
        return getType().isTypeVariable();
    }

    @Override
    public boolean isWildcard() {
        return getType().isWildcard();
    }

    @Override
    public ResolvedArrayType asArrayType() {
        return getType().asArrayType();
    }

    @Override
    public ResolvedReferenceType asReferenceType() {
        return getType().asReferenceType();
    }

    @Override
    public ResolvedTypeParameterDeclaration asTypeParameter() {
        return getType().asTypeParameter();
    }

    @Override
    public ResolvedTypeVariable asTypeVariable() {
        return getType().asTypeVariable();
    }

    @Override
    public ResolvedPrimitiveType asPrimitive() {
        return getType().asPrimitive();
    }

    @Override
    public ResolvedWildcard asWildcard() {
        return getType().asWildcard();
    }

    @Override
    public String describe() {
        return getType().describe();
    }

    @Override
    public ResolvedType replaceTypeVariables(ResolvedTypeParameterDeclaration tp, ResolvedType replaced,
                                             Map<ResolvedTypeParameterDeclaration, ResolvedType> inferredTypes) {
        return getType().replaceTypeVariables(tp, replaced, inferredTypes);
    }

    @Override
    public ResolvedType replaceTypeVariables(ResolvedTypeParameterDeclaration tp, ResolvedType replaced) {
        return getType().replaceTypeVariables(tp, replaced);
    }

    @Override
    public boolean isAssignableBy(ResolvedType other) {
        return getType().isAssignableBy(other);
    }

    @Override
    public boolean equals(Object other) {
        // If is a LazyType then we should find the concrete value of the other type before comparing
        if (other instanceof LazyType) {
            return Objects.equals(getType(), ((LazyType) other).getType());
        } else {
            return Objects.equals(getType(), other);
        }
    }

    @Override
    public int hashCode() {
        return getType().hashCode();
    }

    @Override
    public String toString() {
        return "LazyType{" +
                "concrete=" + concrete +
                '}';
    }

}
