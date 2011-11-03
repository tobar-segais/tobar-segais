/*
 * Copyright 2011 Stephen Connolly
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

package org.tobarsegais.webapp;

import org.apache.commons.io.IOUtils;
import org.apache.lucene.analysis.standard.StandardAnalyzer;
import org.apache.lucene.document.Document;
import org.apache.lucene.document.Field;
import org.apache.lucene.index.IndexWriter;
import org.apache.lucene.index.IndexWriterConfig;
import org.apache.lucene.store.Directory;
import org.apache.lucene.store.RAMDirectory;
import org.apache.lucene.util.Version;
import org.jsoup.Jsoup;
import org.tobarsegais.webapp.data.Entry;
import org.tobarsegais.webapp.data.Extension;
import org.tobarsegais.webapp.data.Plugin;
import org.tobarsegais.webapp.data.Toc;

import javax.servlet.ServletContext;
import javax.servlet.ServletContextEvent;
import javax.servlet.ServletContextListener;
import javax.xml.stream.XMLStreamException;
import java.io.IOException;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.JarURLConnection;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.util.Collections;
import java.util.HashSet;
import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Set;
import java.util.Stack;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

/**
 * Loads all the bundles.
 */
public class ServletContextListenerImpl implements ServletContextListener {
    public void contextInitialized(ServletContextEvent sce) {
        ServletContext application = sce.getServletContext();
        Map<String, Toc> contents = new LinkedHashMap<String, Toc>();
        Directory index = new RAMDirectory();
        IndexWriterConfig indexWriterConfig =
                new IndexWriterConfig(Version.LUCENE_34, new StandardAnalyzer(Version.LUCENE_34));
        IndexWriter indexWriter;
        try {
            indexWriter = new IndexWriter(index, indexWriterConfig);
        } catch (IOException e) {
            application.log("Cannot create indexes. Search will be unavailable.", e);
            indexWriter = null;
        }
        for (String path : (Set<String>) application.getResourcePaths("/WEB-INF/bundles")) {
            if (path.endsWith(".jar")) {
                String key = path.substring("/WEB-INF/bundles/".length(), path.lastIndexOf(".jar"));
                application.log("Parsing " + path);
                URLConnection connection = null;
                try {
                    URL url = new URL("jar:" + application.getResource(path) + "!/");
                    connection = url.openConnection();
                    if (!(connection instanceof JarURLConnection)) {
                        application.log(path + " is not a jar file, ignoring");
                        continue;
                    }
                    JarURLConnection jarConnection = (JarURLConnection) connection;
                    JarFile jarFile = jarConnection.getJarFile();
                    JarEntry pluginEntry = jarFile.getJarEntry("plugin.xml");
                    if (pluginEntry == null) {
                        application.log(path + " does not contain a plugin.xml file, ignoring");
                        continue;
                    }
                    Plugin plugin = Plugin.read(jarFile.getInputStream(pluginEntry));
                    Extension tocExtension = plugin.getExtension("org.eclipse.help.toc");
                    if (tocExtension == null || tocExtension.getFile("toc") == null) {
                        application.log(path + " does not contain a 'org.eclipse.help.toc' extension, ignoring");
                        continue;
                    }
                    JarEntry tocEntry = jarFile.getJarEntry(tocExtension.getFile("toc"));
                    if (tocEntry == null) {
                        application.log(path + " is missing the referenced toc: " + tocExtension.getFile("toc")
                                + ", ignoring");
                        continue;
                    }
                    Toc toc;
                    try {
                        toc = Toc.read(jarFile.getInputStream(tocEntry));
                    } catch (IllegalStateException e) {
                        application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                        continue;
                    }
                    contents.put(key, toc);
                    application.log(path + " successfully parsed and added as " + key);
                    if (indexWriter != null) {
                        application.log("Indexing content of " + path);
                        Set<String> files = new HashSet<String>();
                        Stack<Iterator<? extends Entry>> stack = new Stack<Iterator<? extends Entry>>();
                        stack.push(Collections.singleton(toc).iterator());
                        while (!stack.empty()) {
                            Iterator<? extends Entry> cur = stack.pop();
                            if (cur.hasNext()) {
                                Entry entry = cur.next();
                                stack.push(cur);
                                if (!entry.getChildren().isEmpty()) {
                                    stack.push(entry.getChildren().iterator());
                                }
                                String file = entry.getHref();
                                if (file == null) {
                                    continue;
                                }
                                int hashIndex = file.indexOf('#');
                                if (hashIndex != -1) {
                                    file = file.substring(0, hashIndex);
                                }
                                if (files.contains(file)) {
                                    // already indexed
                                    // todo work out whether to just pull the section
                                    continue;
                                }
                                Document document = new Document();
                                document.add(
                                        new Field("title", entry.getLabel(), Field.Store.YES, Field.Index.ANALYZED));
                                document.add(new Field("href", entry.getHref(), Field.Store.YES, Field.Index.NO));
                                JarEntry docEntry = jarFile.getJarEntry(file);
                                if (docEntry == null) {
                                    // ignore missing file
                                    continue;
                                }
                                InputStream inputStream = null;
                                try {
                                    inputStream = jarFile.getInputStream(docEntry);
                                    org.jsoup.nodes.Document docDoc = Jsoup.parse(IOUtils.toString(inputStream));
                                    document.add(new Field("contents", docDoc.body().text(), Field.Store.NO,
                                            Field.Index.ANALYZED));
                                    indexWriter.addDocument(document);
                                } finally {
                                    IOUtils.closeQuietly(inputStream);
                                }
                            }
                        }
                    }
                } catch (XMLStreamException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } catch (MalformedURLException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } catch (IOException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } finally {
                    if (connection instanceof HttpURLConnection) {
                        // should never be the case, but we should try to be sure
                        ((HttpURLConnection) connection).disconnect();
                    }
                }
            }
        }
        application.setAttribute("toc", Collections.unmodifiableMap(contents));
        application.setAttribute("index", index);
    }

    public void contextDestroyed(ServletContextEvent sce) {
        //To change body of implemented methods use File | Settings | File Templates.
    }
}
