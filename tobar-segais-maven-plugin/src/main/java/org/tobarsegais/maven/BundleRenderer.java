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

import org.apache.maven.doxia.docrenderer.AbstractDocumentRenderer;
import org.apache.maven.doxia.docrenderer.DocumentRenderer;
import org.apache.maven.doxia.docrenderer.DocumentRendererContext;
import org.apache.maven.doxia.docrenderer.DocumentRendererException;
import org.apache.maven.doxia.document.DocumentModel;
import org.apache.maven.doxia.document.DocumentTOC;
import org.apache.maven.doxia.document.DocumentTOCItem;
import org.apache.maven.doxia.module.site.SiteModule;
import org.apache.maven.doxia.module.xhtml.XhtmlSinkFactory;
import org.apache.maven.doxia.sink.SinkFactory;
import org.codehaus.plexus.component.annotations.Component;
import org.codehaus.plexus.component.annotations.Requirement;
import org.codehaus.plexus.util.DirectoryScanner;
import org.codehaus.plexus.util.IOUtil;
import org.codehaus.plexus.util.ReaderFactory;
import org.codehaus.plexus.util.StringUtils;
import org.codehaus.plexus.util.xml.PrettyPrintXMLWriter;
import org.codehaus.plexus.util.xml.XMLWriter;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.jar.JarEntry;
import java.util.jar.JarOutputStream;

/**
 * @author Stephen Connolly
 */
@Component(role = DocumentRenderer.class, hint = "bundle")
public class BundleRenderer extends AbstractDocumentRenderer {

    @Requirement(role = SinkFactory.class, hint = "xhtml")
    private XhtmlSinkFactory sinkFactory;

    @Override
    public void render(Map<String, SiteModule> filesToProcess, File outputDirectory, DocumentModel documentModel)
            throws DocumentRendererException, IOException {
        render(filesToProcess, outputDirectory, documentModel, null);
    }

    public String getOutputExtension() {
        return "jar";
    }

    public void render(Map<String, SiteModule> filesToProcess, File outputDirectory, DocumentModel documentModel,
                       DocumentRendererContext context) throws DocumentRendererException, IOException {
        String outputName = getOutputName(documentModel);

        File outputFile = new File(outputDirectory, outputName + "." + getOutputExtension());
        if (!outputFile.getParentFile().exists()) {
            outputFile.getParentFile().mkdirs();
        }
        FileOutputStream fos = null;
        JarOutputStream jos = null;
        JarXhtmlSink sink = null;
        try {
            fos = new FileOutputStream(outputFile);
            jos = new JarOutputStream(fos);
            sink = new JarXhtmlSink(jos, sinkFactory, (context == null ? ReaderFactory.FILE_ENCODING
                    : context.getInputEncoding()));
            copyResources(sink.getOutputStream());
            if ((documentModel.getToc() == null) || (documentModel.getToc().getItems() == null)) {
                getLogger().info("No TOC is defined in the document descriptor. Merging all documents.");

                renderPluginXml(sink, documentModel, context);
                renderTocXml(sink, documentModel, context);
                renderTocXhtml(sink, documentModel, context);

                mergeAllSources(filesToProcess, sink, context);
            } else {
                getLogger().debug("Using TOC defined in the document descriptor.");

                renderPluginXml(sink, documentModel, context);
                renderTocXml(sink, documentModel, context);
                renderTocXhtml(sink, documentModel, context);

                mergeSourcesFromTOC(documentModel.getToc(), sink, context);
            }


        } finally {
            if (sink != null) {
                sink.close();
            }
            IOUtil.close(jos);
            IOUtil.close(fos);
        }


    }

    /**
     * Copies the contents of the resource directory to an output folder.
     *
     * @param outputStream the destination jar file.
     * @throws java.io.IOException if any.
     */
    protected void copyResources(JarOutputStream outputStream)
            throws IOException {
        File resourcesDirectory = new File(getBaseDir(), "resources");

        if (!resourcesDirectory.isDirectory()) {
            return;
        }

        copyDirectory(resourcesDirectory, outputStream);
    }

    /**
     * Copy content of a directory, excluding scm-specific files.
     *
     * @param source      directory that contains the files and sub-directories to be copied.
     * @param destination destination jar file.
     * @throws java.io.IOException if any.
     */
    protected void copyDirectory(File source, JarOutputStream destination)
            throws IOException {
        if (source.isDirectory()) {
            DirectoryScanner scanner = new DirectoryScanner();

            String[] includedResources = {"**/**"};

            scanner.setIncludes(includedResources);

            scanner.addDefaultExcludes();

            scanner.setBasedir(source);

            scanner.scan();

            List<String> includedFiles = Arrays.asList(scanner.getIncludedFiles());

            for (String name : includedFiles) {
                File sourceFile = new File(source, name);

                destination.putNextEntry(new JarEntry(name));
                FileInputStream input = null;
                try {
                    input = new FileInputStream(sourceFile);
                    IOUtil.copy(input, destination);
                } finally {
                    IOUtil.close(input);
                }
            }
        }
    }

    private void mergeAllSources(Map<String, SiteModule> filesToProcess, JarXhtmlSink sink,
                                 DocumentRendererContext context)
            throws DocumentRendererException, IOException {
        for (Map.Entry<String, SiteModule> entry : filesToProcess.entrySet()) {
            String key = entry.getKey();
            SiteModule module = entry.getValue();
            File fullDoc = new File(getBaseDir(), module.getSourceDirectory() + File.separator + key);
            sink.file(key, fullDoc.lastModified());
            parse(fullDoc.getAbsolutePath(), module.getParserId(), sink, context);
        }
    }

    private void mergeSourcesFromTOC(DocumentTOC toc, JarXhtmlSink sink, DocumentRendererContext context)
            throws IOException, DocumentRendererException {
        parseTocItems(toc.getItems(), sink, context);
    }

    private void parseTocItems(List<DocumentTOCItem> items, JarXhtmlSink sink, DocumentRendererContext context)
            throws IOException, DocumentRendererException {
        for (DocumentTOCItem tocItem : items) {
            if (tocItem.getRef() == null) {
                if (getLogger().isInfoEnabled()) {
                    getLogger().info("No ref defined for tocItem " + tocItem.getName());
                }

                continue;
            }

            String href = StringUtils.replace(tocItem.getRef(), "\\", "/");
            if (href.lastIndexOf('.') != -1) {
                href = href.substring(0, href.lastIndexOf('.'));
            }

            renderModules(href, sink, tocItem, context);

            if (tocItem.getItems() != null) {
                parseTocItems(tocItem.getItems(), sink, context);
            }
        }
    }

    private void renderModules(String href, JarXhtmlSink sink, DocumentTOCItem tocItem,
                               DocumentRendererContext context)
            throws DocumentRendererException, IOException {
        Collection<SiteModule> modules = siteModuleManager.getSiteModules();
        for (SiteModule module : modules) {
            File moduleBasedir = new File(getBaseDir(), module.getSourceDirectory());

            if (moduleBasedir.exists()) {
                String doc = href + "." + module.getExtension();
                File source = new File(moduleBasedir, doc);

                // Velocity file?
                if (!source.exists()) {
                    if (href.indexOf("." + module.getExtension()) != -1) {
                        doc = href + ".vm";
                    } else {
                        doc = href + "." + module.getExtension() + ".vm";
                    }
                    source = new File(moduleBasedir, doc);
                }

                if (source.exists()) {
                    sink.file(tocItem.getRef(), source.lastModified());
                    parse(source.getPath(), module.getParserId(), sink, context);
                }
            }
        }
    }

    private void renderTocXml(JarXhtmlSink sink, DocumentModel model, DocumentRendererContext context)
            throws IOException {
        final JarOutputStream outputStream = sink.getOutputStream();
        outputStream.putNextEntry(new JarEntry("toc.xml"));
        PrintWriter pw = new PrintWriter(outputStream);
        try {
            XMLWriter w = new PrettyPrintXMLWriter(pw, "  ", "\n", "UTF-8", null);
            w.startElement("toc");
            w.addAttribute("label", model.getCover().getCoverTitle());
            w.addAttribute("href", "_toc.html");
            writeTocItems(w, model.getToc().getItems());
            w.endElement();
        } finally {
            pw.flush();
        }
    }

    private void writeTocItems(XMLWriter w, List<DocumentTOCItem> items) {
        if (items != null) {
            for (DocumentTOCItem item : items) {
                w.startElement("topic");
                w.addAttribute("label", item.getName());
                w.addAttribute("href", item.getRef() + ".html");
                writeTocItems(w, item.getItems());
                w.endElement();
            }
        }
    }

    private void renderTocXhtml(JarXhtmlSink sink, DocumentModel model, DocumentRendererContext context)
            throws IOException {
        sink.file("_toc");
        sink.head();
        sink.title();
        sink.text(model.getToc().getName());
        sink.title_();
        sink.head_();
        sink.body();
        sink.toc(model.getToc());
        sink.body_();
        sink.file_();
    }

    private void renderPluginXml(JarXhtmlSink sink, DocumentModel model, DocumentRendererContext context)
            throws IOException {
        final JarOutputStream outputStream = sink.getOutputStream();
        outputStream.putNextEntry(new JarEntry("plugin.xml"));
        PrintWriter pw = new PrintWriter(outputStream);
        try {
            XMLWriter w = new PrettyPrintXMLWriter(pw, "  ", "\n", "UTF-8", null);
            w.startElement("plugin");
            w.addAttribute("name", model.getCover().getCoverTitle());
            w.addAttribute("id", ""); // TODO
            w.addAttribute("version", model.getCover().getCoverVersion());
            w.addAttribute("provider-name", StringUtils.defaultString(model.getCover().getCompanyName()));
            w.startElement("extension");
            w.addAttribute("point", "org.eclipse.help.toc");
            w.startElement("toc");
            w.addAttribute("file", "toc.xml");
            w.addAttribute("primary", "true");
            w.endElement();
            w.endElement();
            w.endElement();
        } finally {
            pw.flush();
        }
    }


}
