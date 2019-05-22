pipeline {
    agent {
        node {
            label 'operator-sdk'
        }
    }

    libraries {
        lib('fh-pipeline-library')
    }
    
    environment {
        GOPATH = "${env.WORKSPACE}/"
        PATH = "${env.PATH}:${env.WORKSPACE}/bin"
        GOOS = "linux"
        GOARCH = "amd64"
        CGO_ENABLED = 0
        OPERATOR_NAME = "unifiedpush-operator"
        OPERATOR_TEST_NAME = "${env.OPERATOR_NAME}-test"
        OPENSHIFT_PROJECT_NAME = "test-${env.OPERATOR_NAME}-${currentBuild.number}-${currentBuild.startTimeInMillis}"
        CLONED_REPOSITORY_PATH = "src/github.com/aerogear/unifiedpush-operator"
        OPERATOR_CONTAINER_IMAGE_NAME = "quay.io/aerogear/${env.OPERATOR_NAME}:${env.BRANCH_NAME}"
        OPERATOR_TEST_CONTAINER_IMAGE_NAME = "docker-registry.default.svc:5000/${env.OPENSHIFT_PROJECT_NAME}/${env.OPERATOR_TEST_NAME}:latest"
    }
    
    options {
        checkoutToSubdirectory("src/github.com/aerogear/unifiedpush-operator")
    }

    stages {
        stage("Trust"){
            steps{
                enforceTrustedApproval('aerogear')
            }
            post{
                failure{
                    echo "====++++'Trust' execution failed++++===="
                    echo "You are not authorized to run this job"
                }
        
            }
        }

        stage("Create an OpenShift project") {
            steps {
                script {
                    openshift.withCluster('operators-test-cluster') {
                        generateKubeConfig()
                        openshift.newProject(env.OPENSHIFT_PROJECT_NAME)
                    }
                }
            }
        }

        stage("Build code binary"){
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        sh """
                        make code/compile
                        """
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build code binary' execution failed++++===="
                    echo "Try to run 'make code/compile' locally and make sure it pass"
                }
            }
        }
        stage("Build & push container image") {
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        withCredentials([usernamePassword(credentialsId: 'quay-aerogear-bot', usernameVariable: 'quayUsername', passwordVariable: "quayPassword")]) {
                            final String buildConfig = """
                            {
                                "apiVersion": "build.openshift.io/v1",
                                "kind": "BuildConfig",
                                "metadata": {
                                    "labels": {
                                    "build": "${env.OPERATOR_NAME}"
                                    },
                                    "name": "${env.OPERATOR_NAME}"
                                },
                                "spec": {
                                    "runPolicy": "Serial",
                                    "failedBuildsHistoryLimit": 1,
                                    "successfulBuildsHistoryLimit": 1,
                                    "source": {
                                        "binary": {},
                                        "type": "Binary"
                                    },
                                    "strategy": {
                                        "dockerStrategy": {
                                            "dockerfilePath": "build/Dockerfile"
                                        },
                                        "type": "Docker"
                                        },
                                    "output": {
                                        "to": {
                                            "kind": "DockerImage",
                                            "name": "${env.OPERATOR_CONTAINER_IMAGE_NAME}"
                                        },
                                        "pushSecret": {
                                            "name": "quay-bot"
                                        }
                                    }
                                }
                            }
                            """
                            openshift.withCluster('operators-test-cluster') {
                                openshift.withProject(env.OPENSHIFT_PROJECT_NAME) {
                                    openshift.create(
                                        "secret", "docker-registry", "quay-bot",
                                        "--docker-username=${quayUsername}",
                                        "--docker-password=${quayPassword}", "--docker-server=quay.io"
                                        )
                                    openshift.apply buildConfig
                                    def build = openshift.startBuild("${env.OPERATOR_NAME}", "--from-dir=.")

                                    waitUntil {
                                        build.object().status.phase == "Running"
                                    }
                                    build.logs('-f')
                                }
                            }
                        }
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build & push container image' execution failed++++===="
                }
            }
        }
        stage("Build test binary"){
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        sh """
                        make test/compile
                        """
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build test binary' execution failed++++===="
                    echo "Try to run 'make test/compile' locally and make sure it pass"
                }
            }
        }
        stage("Build operator-test image") {
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        def operatorTestDockerfileContent = """
                        FROM ${env.OPERATOR_CONTAINER_IMAGE_NAME}
                        ADD build/_output/bin/unifiedpush-operator-test /usr/local/bin/unifiedpush-operator-test
                        ADD deploy/operator.yaml /namespaced.yaml
                        ADD build/test-framework/go-test.sh /go-test.sh
                        """
                        openshift.withCluster('operators-test-cluster') {
                            openshift.withProject(env.OPENSHIFT_PROJECT_NAME) {
                                writeFile file: "Dockerfile", text: "${operatorTestDockerfileContent}"
                                sh "yq w -i deploy/operator.yaml spec.template.spec.containers[0].image ${env.OPERATOR_TEST_CONTAINER_IMAGE_NAME}"
                                openshift.newBuild("--name=${env.OPERATOR_TEST_NAME}", "--binary")
                                def build = openshift.startBuild("${env.OPERATOR_TEST_NAME}", "--from-dir=.")
                                waitUntil {
                                    build.object().status.phase == "Running"
                                }
                                build.logs('-f')
                            }
                        }
                    }
                }
            }
            post{
                failure{
                    echo "====++++Build operator-test image execution failed++++===="
                }
        
            }
        }
        stage("Test operator") {
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        openshift.withCluster('operators-test-cluster') {
                            openshift.withProject(env.OPENSHIFT_PROJECT_NAME) {
                                sh """
                                make NAMESPACE=${env.OPENSHIFT_PROJECT_NAME} cluster/prepare
                                operator-sdk test cluster ${env.OPERATOR_TEST_CONTAINER_IMAGE_NAME} --namespace ${env.OPENSHIFT_PROJECT_NAME} --service-account ${env.OPERATOR_NAME}
                                """
                            }
                        }
                    }
                }
            }
            post{
                failure{
                    echo "====++++Test operator execution failed++++===="
                }
            }
        }
    }
    post {
        always{
            script {
                openshift.withCluster('operators-test-cluster') {
                    openshift.delete("project", env.OPENSHIFT_PROJECT_NAME)
                }
            }
        }
    }
}