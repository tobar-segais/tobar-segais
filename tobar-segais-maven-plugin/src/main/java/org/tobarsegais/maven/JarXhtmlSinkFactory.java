/*
 * Copyright 2013 Stephen Connolly
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.tobarsegais.maven;

import org.apache.maven.doxia.module.xhtml.XhtmlSinkFactory;
import org.apache.maven.doxia.sink.Sink;
import org.apache.maven.doxia.sink.SinkFactory;
import org.codehaus.plexus.component.annotations.Component;
import org.codehaus.plexus.component.annotations.Requirement;
import org.codehaus.plexus.util.WriterFactory;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.util.jar.JarOutputStream;

/**
 * @author Stephen Connolly
 */
@Component(role = SinkFactory.class, hint = "bundle")
public class JarXhtmlSinkFactory implements SinkFactory {

    @Requirement(role = SinkFactory.class, hint = "xhtml")
    private XhtmlSinkFactory factory;

    public Sink createSink(File outputDir, String outputName) throws IOException {
        return createSink(outputDir, outputName, WriterFactory.UTF_8);
    }

    public Sink createSink(File outputDir, String outputName, String encoding) throws IOException {
        return createSink(new FileOutputStream(new File(outputDir, outputName)), encoding);
    }

    public Sink createSink(OutputStream out) throws IOException {
        return createSink(out, WriterFactory.UTF_8);
    }

    public Sink createSink(OutputStream out, String encoding) throws IOException {
        return new JarXhtmlSink(new JarOutputStream(out), factory, encoding);
    }
}
