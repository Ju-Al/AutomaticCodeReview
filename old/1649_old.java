/*
 * Copyright (C) 2007-2010 Júlio Vilmar Gesser.
 * Copyright (C) 2011, 2013-2016 The JavaParser Team.
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

package com.github.javaparser.ast.comments;

import com.github.javaparser.JavaParser;
import com.github.javaparser.ast.CompilationUnit;
import com.github.javaparser.ast.Modifier;
import com.github.javaparser.ast.body.ClassOrInterfaceDeclaration;
import com.github.javaparser.ast.body.MethodDeclaration;
import com.github.javaparser.ast.expr.Name;
import com.github.javaparser.javadoc.Javadoc;
import com.github.javaparser.javadoc.description.JavadocDescription;
import com.github.javaparser.printer.PrettyPrinterConfiguration;
import org.junit.Test;

import java.util.EnumSet;

import static com.github.javaparser.JavaParser.parse;
import static com.github.javaparser.JavaParser.parseName;
import static com.github.javaparser.utils.Utils.EOL;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

public class CommentTest {

    private static final PrettyPrinterConfiguration prettyPrintConfig = new PrettyPrinterConfiguration().setIndentSize(2);

    @Test
    public void removeOrphanComment() {
        ClassOrInterfaceDeclaration decl = new ClassOrInterfaceDeclaration(EnumSet.noneOf(Modifier.class), false, "A");
        Comment c = new LineComment("A comment");
        decl.addOrphanComment(c);
        assertEquals(1, decl.getOrphanComments().size());
        assertTrue(c.remove());
        assertEquals(0, decl.getOrphanComments().size());
    }

    @Test
    public void removeAssociatedComment() {
        ClassOrInterfaceDeclaration decl = new ClassOrInterfaceDeclaration(EnumSet.noneOf(Modifier.class), false, "A");
        Comment c = new LineComment("A comment");
        decl.setComment(c);
        assertEquals(true, decl.getComment().isPresent());
        assertTrue(c.remove());
        assertEquals(false, decl.getComment().isPresent());
    }

    @Test
    public void cannotRemoveCommentNotUsedAnywhere() {
        Comment c = new LineComment("A comment");
        assertFalse(c.remove());
    }

    @Test
    public void unicodeEscapesArePreservedInComments() {
        CompilationUnit cu = parse("// xxx\\u2122xxx");
        Comment comment = cu.getAllContainedComments().get(0);
        assertEquals(" xxx\\u2122xxx", comment.getContent());
    }

    @Test
    public void testReplaceDuplicateJavaDocComment() {
        // Arrange
        CompilationUnit cu = parse("public class MyClass {" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL);

        MethodDeclaration methodDeclaration = cu.findFirst(MethodDeclaration.class).get();

        // Act
        Javadoc javadoc = new Javadoc(JavadocDescription.parseText("Change Javadoc"));
        methodDeclaration.setJavadocComment("", javadoc);

        // Assert
        assertEquals("public class MyClass {" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Change Javadoc" + EOL +
                "   */" + EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL, cu.toString(prettyPrintConfig));
    }

    @Test
    public void testRemoveDuplicateComment() {
        // Arrange
        CompilationUnit cu = parse("public class MyClass {" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL);

        MethodDeclaration methodDeclaration = cu.findFirst(MethodDeclaration.class).get();

        // Act
        methodDeclaration.removeComment();

        // Assert
        assertEquals("public class MyClass {" + EOL +
                EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL, cu.toString(prettyPrintConfig));
    }

    @Test
    public void testRemoveDuplicateJavaDocComment() {
        // Arrange
        CompilationUnit cu = parse("public class MyClass {" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL);

        MethodDeclaration methodDeclaration = cu.findAll(MethodDeclaration.class).get(1);

        // Act
        methodDeclaration.removeJavaDocComment();

        // Assert
        assertEquals("public class MyClass {" + EOL +
                EOL +
                "  /**" + EOL +
                "   * Comment A" + EOL +
                "   */" + EOL +
                "  public void oneMethod() {" + EOL +
                "  }" + EOL +
                EOL +
                "  public void anotherMethod() {" + EOL +
                "  }" + EOL +
                "}" +
                EOL, cu.toString(prettyPrintConfig));
    }
}
