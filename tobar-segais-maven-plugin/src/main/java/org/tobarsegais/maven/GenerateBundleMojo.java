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

import org.apache.maven.doxia.docrenderer.DocumentRenderer;
import org.apache.maven.doxia.docrenderer.DocumentRendererException;
import org.apache.maven.doxia.document.DocumentModel;
import org.apache.maven.plugins.annotations.Component;
import org.apache.maven.plugins.annotations.LifecyclePhase;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;

import java.io.File;
import java.io.IOException;

/**
 * @author Stephen Connolly
 */
@Mojo(name = "bundle", defaultPhase = LifecyclePhase.GENERATE_RESOURCES, threadSafe = true)
public class GenerateBundleMojo extends AbstractGenerateMojo {

    @Parameter(defaultValue = "${project.build.directory}/${project.build.finalName}-docs.jar")
    private File outputFile;
    @Component(role = DocumentRenderer.class, hint = "bundle")
    private BundleRenderer renderer;

    protected void render(DocumentModel model) throws DocumentRendererException, IOException {
        renderer.render(sourceDirectory, outputFile.getParentFile(), model);
        projectHelper.attachArtifact(project, "pdf", "docs", outputFile);
    }

    public File getOutputFile() {
        return outputFile;
    }

    public BundleRenderer getRenderer() {
        return renderer;
    }
}
