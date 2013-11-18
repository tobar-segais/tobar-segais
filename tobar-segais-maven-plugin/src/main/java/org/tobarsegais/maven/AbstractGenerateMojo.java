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

import org.apache.commons.io.FilenameUtils;
import org.apache.maven.doxia.Doxia;
import org.apache.maven.doxia.docrenderer.AbstractDocumentRenderer;
import org.apache.maven.doxia.docrenderer.DocumentRenderer;
import org.apache.maven.doxia.docrenderer.DocumentRendererException;
import org.apache.maven.doxia.document.DocumentAuthor;
import org.apache.maven.doxia.document.DocumentCover;
import org.apache.maven.doxia.document.DocumentMeta;
import org.apache.maven.doxia.document.DocumentModel;
import org.apache.maven.doxia.document.DocumentTOC;
import org.apache.maven.doxia.document.DocumentTOCItem;
import org.apache.maven.doxia.module.site.SiteModule;
import org.apache.maven.doxia.parser.ParseException;
import org.apache.maven.doxia.parser.Parser;
import org.apache.maven.doxia.parser.manager.ParserNotFoundException;
import org.apache.maven.model.Developer;
import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.Component;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;
import org.apache.maven.project.MavenProjectHelper;
import org.codehaus.plexus.util.IOUtil;
import org.codehaus.plexus.util.StringUtils;
import org.codehaus.plexus.util.xml.pull.XmlPullParserException;
import org.tobarsegais.maven.model.BundleFile;
import org.tobarsegais.maven.model.BundleModel;
import org.tobarsegais.maven.model.io.xpp3.BundleXpp3Reader;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Map;

public abstract class AbstractGenerateMojo extends AbstractMojo {
    @Parameter(defaultValue = "${basedir/src/docs/bundle.xml}")
    protected File bundle;
    @Parameter(defaultValue = "${basedir}/src/docs")
    protected File sourceDirectory;
    @Parameter(defaultValue = "${project.build.outputDirectory}")
    protected File outputDirectory;
    /**
     * Skip compilation.
     */
    @Parameter(property = "tobairsegais.skip", defaultValue = "false")
    protected boolean skip;
    /**
     * The character encoding scheme to be applied when reading source files.
     */
    @Parameter(defaultValue = "${project.build.sourceEncoding}")
    protected String sourceEncoding;
    /**
     * The character encoding scheme to be applied when writing output files.
     */
    @Parameter(defaultValue = "${project.build.outputEncoding}")
    protected String outputEncoding;
    @Component
    protected MavenProject project;
    @Component
    protected MavenProjectHelper projectHelper;
    @Component
    protected Doxia doxia;

    public DocumentCover getDocumentCover(BundleModel bundleModel, Date date) {
        DocumentCover cover = new DocumentCover();
        cover.setAuthors(getAuthors());
        cover.setCompanyName(StringUtils.defaultString(bundleModel.getCompanyName(),
                project.getOrganization() != null && StringUtils.isNotBlank(project.getOrganization().getName())
                        ? project
                        .getOrganization().getName()
                        : null));
        cover.setCompanyLogo(
                bundleModel.getCompanyLogo() != null ? new File(sourceDirectory, bundleModel.getCompanyLogo())
                        .getAbsolutePath() : null);
        cover.setProjectName(bundleModel.getProjectName());
        cover.setProjectLogo(
                bundleModel.getProjectLogo() != null ? new File(sourceDirectory, bundleModel.getProjectLogo())
                        .getAbsolutePath() : null);
        cover.setCoverDate(date);
        cover.setCoverVersion(StringUtils.defaultString(bundleModel.getVersion(), project.getVersion()));
        cover.setCoverSubTitle(bundleModel.getSubject());
        cover.setCoverTitle(StringUtils.defaultString(bundleModel.getTitle(), getProjectName()));
        return cover;
    }

    public DocumentMeta getDocumentMeta(BundleModel bundleModel, Date date) {
        DocumentMeta meta = new DocumentMeta();
        meta.setAuthors(getAuthors());
        meta.setCreationDate(date);
        meta.setCreator(System.getProperty("user.name"));
        meta.setDate(date);
        meta.setDescription(StringUtils.defaultString(bundleModel.getDescription(), project.getDescription()));
        meta.setInitialCreator(System.getProperty("user.name"));
        meta.setSubject(bundleModel.getSubject());
        meta.setTitle(StringUtils.defaultString(bundleModel.getTitle(), getProjectName()));
        return meta;
    }

    public String getProjectName() {
        return StringUtils.isEmpty(project.getName())
                ? project.getGroupId() + ":" + project.getArtifactId()
                : project.getName();
    }

    List<DocumentAuthor> getAuthors() {
        if (project.getDevelopers() == null) {
            return null;
        }
        List<DocumentAuthor> authors = new ArrayList<DocumentAuthor>(project.getDevelopers().size());
        for (Developer d : (List<Developer>) project.getDevelopers()) {
            DocumentAuthor author = new DocumentAuthor();
            author.setName(d.getName());
            author.setCompanyName(d.getOrganization());
            authors.add(author);
        }
        return authors;
    }

    public void execute() throws MojoExecutionException, MojoFailureException {
        if (skip) {
            getLog().info("Generation skipped");
            return;
        }
        if (!bundle.isFile()) {
            getLog().warn("Missing bundle descriptor '" + bundle + "'. Nothing to do.");
            return;
        }
        if (sourceDirectory.isFile()) {
            throw new MojoExecutionException("Source directory '" + sourceDirectory + "' is not a directory");
        }
        if (!sourceDirectory.isDirectory()) {
            getLog().info("Source directory '" + sourceDirectory + " does not exist. Nothing to do.");
            return;
        }
        if (outputDirectory.isFile()) {
            throw new MojoExecutionException("Output directory '" + outputDirectory + "' is not a directory");
        }
        if (!outputDirectory.isDirectory()) {
            if (!outputDirectory.mkdirs()) {
                throw new MojoExecutionException("Could not create output directory '" + outputDirectory + "'.");
            }
        }
        if (!getOutputFile().getParentFile().isDirectory()) {
            if (!getOutputFile().getParentFile().mkdirs()) {
                throw new MojoExecutionException(
                        "Could not create output directory '" + getOutputFile().getParentFile() + "'.");
            }
        }
        BundleModel bundleModel;
        BundleXpp3Reader reader = new BundleXpp3Reader();
        FileInputStream fis = null;
        try {
            fis = new FileInputStream(bundle);
            bundleModel = reader.read(fis, false);
        } catch (XmlPullParserException e) {
            throw new MojoExecutionException("Could not read bundle '" + bundle + "'", e);
        } catch (IOException e) {
            throw new MojoExecutionException("Could not read bundle '" + bundle + "'", e);
        } finally {
            IOUtil.close(fis);
        }

        final Date date = new Date();
        try {
            final DocumentModel model = new DocumentModel();
            model.setModelEncoding(StringUtils.defaultString(outputEncoding, "UTF-8"));
            model.setOutputName(getOutputFile().getName());
            model.setMeta(getDocumentMeta(bundleModel, date));
            model.setCover(getDocumentCover(bundleModel, date));
            model.setToc(new DocumentTOC());
            model.getToc().setDepth(3);
            model.getToc().setName("Contents");
            final Map<String, SiteModule> toProcess = getRenderer().getFilesToProcess(sourceDirectory);
            for (BundleFile f : bundleModel.getFiles()) {
                final DocumentTOCItem item = new DocumentTOCItem();
                if (StringUtils.isBlank(f.getTitle())) {
                    boolean found = false;
                    for (Map.Entry<String, SiteModule> entry : toProcess.entrySet()) {
                        if (f.getSrc().equals(entry.getKey())
                                || f.getSrc().equals(FilenameUtils.removeExtension(entry.getKey()))) {
                            try {
                                final Parser parser = doxia
                                        .getParser(entry.getValue().getParserId());
                                fis = null;
                                CapturingSink sink = null;
                                try {
                                    fis = new FileInputStream(
                                            new File(new File(sourceDirectory, entry.getValue().getSourceDirectory()),
                                                    entry.getKey()));
                                    sink = new CapturingSink();
                                    parser.parse(new InputStreamReader(fis, sourceEncoding), sink);
                                } catch (ParseException e) {
                                    getLog().error(e);
                                } finally {
                                    if (sink != null) {
                                        sink.close();
                                    }
                                    IOUtil.close(fis);
                                }
                                item.setName(sink.getTitle());
                                found = true;
                            } catch (ParserNotFoundException e) {
                                e.printStackTrace();  //To change body of catch statement use File | Settings | File
                                // Templates.
                            }
                            break;
                        }
                    }
                    if (!found) {
                        item.setName(new File(sourceDirectory, f.getSrc()).getName());
                    }

                } else {
                    item.setName(f.getTitle());
                }
                item.setRef(f.getSrc()+".html");
                model.getToc().addItem(item);
            }

            render(model);
        } catch (DocumentRendererException e) {
            throw new MojoExecutionException(e.getMessage(), e);
        } catch (IOException e) {
            throw new MojoExecutionException(e.getMessage(), e);
        }
    }

    protected abstract AbstractDocumentRenderer getRenderer();

    protected abstract File getOutputFile();

    protected abstract void render(DocumentModel model) throws DocumentRendererException, IOException;

}