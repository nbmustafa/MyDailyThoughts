# AWS Certified Machine Learning Engineer Associate - MLA-C01 Exam - Study Notes

## Important Links

- **📝 Official Exam Guide (PDF)**  
  [AWS Certified Machine Learning Engineer – Associate (MLA-C01) Official Guide](https://d1.awsstatic.com/training-and-certification/docs-ml/ml-engineer-associate-exam-guide.pdf)

- **📚 AWS Certified Machine Learning Engineer – Associate Page**  
  [Exam Overview, Registration, Sample Questions](https://aws.amazon.com/certification/certified-machine-learning-engineer-associate/)

- **🔧 SageMaker Developer Guide**  
  [Amazon SageMaker Documentation](https://docs.aws.amazon.com/sagemaker/latest/dg/whatis.html)

- **📊 AWS Well-Architected ML Lens**  
  [Machine Learning Lens Whitepaper](https://docs.aws.amazon.com/wellarchitected/latest/machine-learning-lens/welcome.html)

## AWS MLA-C01 Exam – Community Insights & Cheat Sheet

### Study Tips & Focus Areas  
- **End-to-End ML Pipeline:** Be prepared for **full lifecycle** coverage – data prep, feature engineering, model training, deployment, and monitoring. The exam demands broad ML knowledge beyond just AWS services.  
- **Hands-On Practice:** Emulate real scenarios with **Amazon SageMaker** and data services. The exam is *heavy on SageMaker* (e.g. building, tuning, deploying models), so practical experience using its features greatly helps.  
- **Breadth Over Depth:** Focus on understanding **use-cases and best practices** for AWS ML tools rather than deep math. *Know what SageMaker’s built-in algorithms are used for* (e.g. XGBoost for regression/classification, BlazingText for word embeddings) but you won’t need to derive their formulas.  
- **Leverage Practice Exams:** Use community practice tests to identify weak areas, but note that the real exam can be more complex. Test-takers found some practice exams easier than the actual questions, so ensure you truly grasp concepts instead of memorizing answers.

### High-Value AWS Services & Tools  
- **Amazon SageMaker Ecosystem:** Expect many questions on SageMaker. Know its **capabilities**: Experiments & Lineage tracking, Pipelines (for ML workflows), AutoPilot (auto ML), Model Monitor (drift detection), Model Cards (documentation), Data Wrangler (data prep), and Clarify (bias/explainability).  
- **Data Ingestion & Storage:** Be familiar with data lakes and ETL on AWS. **S3** is fundamental for data storage, but also know **AWS Glue** (ETL jobs, DataBrew for no-code data prep, Glue Catalog), **Kinesis** (streaming ingestion), and even using **Amazon FSx for Lustre** for high-speed training data storage. Lake Formation can appear in context of data governance.  
- **Analytics & Querying:** Understand tools to analyze data in S3. **Amazon Athena** (serverless SQL querying) is frequently used for analytics on S3 data. Columnar formats like **Parquet** are often preferred – efficient storage & faster queries for Athena/EMR (vs. CSV/JSON).  
- **AI/ML Managed Services:** Know the use-cases of AI services like **Amazon Rekognition** (image analysis), **Comprehend** (NLP text analysis), **Polly** (text-to-speech), **Textract** (OCR), etc. These appear in questions about choosing the right service for a given ML task.  
  - e.g. *PII in documents -> Comprehend or Macie*  
  - e.g. *Transcribe audio -> Amazon Transcribe*  
- **Serverless & Orchestration:** Many scenarios favor serverless architectures. Know **AWS Lambda** (for ETL or inference triggers) and **AWS Step Functions** vs. SageMaker Pipelines for orchestrating ML workflows. AWS Batch can come up for large-scale ML jobs.  
  - Default to managed/serverless solutions when asked for the *easiest scalable* approach.  
- **Security & Compliance:** Understand how to secure ML environments.  
  - IAM roles: SageMaker execution roles, least privilege  
  - Networking: VPC endpoints, security groups for SageMaker  
  - **Sensitive data (PII) -> Amazon Macie**  
  - Encryption: KMS  
  - Auditing/logging: CloudTrail, CloudWatch  

### ML Theory & Key Concepts Tested  
- **Model Evaluation Metrics:**  
  - Classification: Accuracy, Precision, Recall, F1 score  
  - Imbalanced data: Prefer F1 / Precision / Recall over Accuracy  
  - Regression: RMSE, MAE, MSE  
  - ROC-AUC: For binary classification  
- **Overfitting vs Underfitting:**  
  - Overfit = High train acc, low test acc → regularize, simplify model  
  - Underfit = Low acc on both → more complex model, more features  
  - Avoid **data leakage** (test data in train set)  
- **Bias & Fairness:**  
  - Bias mitigation & explainability = **SageMaker Clarify**  
  - Understand what fairness, variance, and bias mean in ML  
- **MLOps & CI/CD:**  
  - Know CI/CD integration with SageMaker Pipelines  
  - Model registry, versioning, automation with CodePipeline/CodeBuild  
- **Data Formats & Featurization:**  
  - **Parquet/ORC = efficient columnar formats** → best for analytics  
  - **RecordIO = used in SageMaker training**  
  - One-hot encoding for categoricals  
  - Scaling/normalizing numerical values  
  - Handling class imbalance: oversampling, class weights  
- **Misc. Algorithm Concepts:**  
  - Know basic use-cases:  
    - Linear/Logistic Regression → Regression/Classification  
    - Clustering vs Classification  
    - XGBoost, LightGBM → Boosting algorithms  
  - When ML is *not* needed: use heuristic if sufficient  

### Question Patterns & Surprises  
- **Scenario-Based Questions:**  
  - Focus on keywords like: *“lowest latency”*, *“most cost-effective”*, *“highest accuracy”*  
  - Questions often about choosing the **best AWS tool/approach** for a situation  
- **Best Practices Emphasis:**  
  - Prefer **serverless or managed services**  
  - Proper monitoring: CloudWatch, SageMaker Model Monitor  
  - Think *scalable, secure, low-maintenance*  
- **New Question Formats:**  
  - Matching (e.g. use-case -> service)  
  - Ordering (put steps in correct sequence)  
  - Case-study style: multiple Qs per scenario  
- **Under-Emphasized Topics:**  
  - **Generative AI / Bedrock** is light-touch – mostly high-level  
  - No deep dives into transformer internals or GPT mechanics  
  - No need for math derivations or Python syntax  
- **Unexpected Services:**  
  - **Amazon FSx for Lustre** – high-speed training storage  
  - **AWS Lake Formation** – permission management for data lakes  
  - **Lookout for Equipment / Metrics / Vision**, **CodeGuru** – know basic purpose only  

## Model Evaluation Metrics

---

### Classification Metrics

#### Accuracy
Use when **class distribution is balanced** and both error types matter.  
*Example: Manufacturing quality control.*  
Formula: `(TP + TN) / (TP + TN + FP + FN)`

#### Precision
Use when **false positives are more costly**.  
*Example: Flagging non-spam as spam.*  
Formula: `TP / (TP + FP)`

#### Recall (Sensitivity)
Use when **false negatives are more costly**.  
*Example: Missing cancer diagnosis, fraud detection.*  
Formula: `TP / (TP + FN)`

#### Specificity
Focuses on identifying **true negatives**, minimizes false positives.  
*Example: Avoiding false alarms in healthy patients.*  
Formula: `TN / (TN + FP)`

#### F1 Score
Harmonic mean of precision and recall — use for **imbalanced datasets**.  
*Example: Fraud detection.*  
Formula: `2 * (Precision * Recall) / (Precision + Recall)`

#### AUC-ROC (Area Under the Curve - Receiver Operating Characteristic)
Measures the model’s ability to distinguish between classes at various thresholds.  
- Higher AUC = better classifier.  
- Useful when you want to evaluate performance across different thresholds.  
*Example: Medical testing models.*  
Range: `0.0 – 1.0` (closer to 1 = better)

---

### Regression Metrics

#### Mean Absolute Error (MAE)
Average of absolute errors between predictions and true values.  
*Example: House price prediction.*  
Formula: `(1/n) * Σ|Actual - Predicted|`

#### Mean Squared Error (MSE)
Penalizes larger errors more heavily than MAE.  
*Example: Forecasting stock prices.*  
Formula: `(1/n) * Σ(Actual - Predicted)^2`

#### Root Mean Squared Error (RMSE)
Square root of MSE — same units as target variable.  
*Example: Temperature prediction.*  
Formula: `√MSE`

#### R-Squared (Coefficient of Determination)
Represents variance explained by the model.  
*Example: Linear regression evaluation.*  
Formula: `1 - (Σ(Actual - Predicted)^2 / Σ(Actual - Mean)^2)`

#### Adjusted R-Squared
Modified R² that adjusts for number of features.  
Useful when comparing models with different number of predictors.  
Formula: `1 - [(1 - R²) * (n - 1) / (n - p - 1)]`

---

### NLP-Specific Metrics

#### Perplexity
Used to evaluate **language models** (e.g., GPT, LLMs).  
Measures how well a model predicts a sequence. Lower is better.  
*Example: Evaluating next-word prediction models.*  
Formula: `Perplexity = 2^CrossEntropy`

#### BLEU (Bilingual Evaluation Understudy Score)
Used for **evaluating machine translation** quality.  
Compares predicted translation to one or more reference translations.  
Score range: `0 (worst)` to `1 (perfect match)`.  
*Example: Comparing model-generated translations to human-translated sentences.*

---

### Visualization Tools

#### Confusion Matrix
Used for classification. Shows **TP, FP, FN, TN**.  
→ Helps analyze which classes are misclassified.

#### ROC Curve
Used for **binary classification** to show the trade-off between **TPR (Recall)** and **FPR** at different thresholds.  
→ Curve closer to top-left is better.

#### Precision-Recall Curve
Alternative to ROC when **classes are imbalanced**.  
→ Shows trade-off between precision and recall at various thresholds.

#### Heat Maps
Visualize correlation between features or confusion matrices.  
→ Helpful in identifying **multicollinearity** or misclassifications in multiclass problems.


---

## Modeling Approaches

### Logistic Regression
For **binary classification**.  
*Example: Customer churn prediction.*

### Linear Regression
For **predicting continuous values**.  
*Example: Predicting home prices.*

### Multiclass Classification
For multi-label targets. Use **Softmax**, **Multinomial Logistic**, **XGBoost**.  
*Example: Classify product category.*

### Ordinal Regression
Use when target classes have order.  
*Example: Ratings — low, medium, high.*

### Time-Series Forecasting
Use **ARIMA**, **DeepAR**, **Prophet**.  
*Example: Stock price forecasting.*

### Ensemble Methods
Combine multiple models.  
- **Bagging (Random Forest)**: Reduces variance  
- **Boosting (XGBoost, LightGBM)**: Reduces bias

### Neural Networks
- **MLP**: General-purpose (tabular data)  
- **CNN**: Images  
- **RNN/LSTM**: Sequences, time-series

### Autoencoders
Used for **dimensionality reduction**, **anomaly detection**.

### Naive Bayes
Probabilistic model based on Bayes’ Theorem.  
Assumes feature independence.  
*Example: Spam email classification.*

---

## Data Prep & Feature Engineering

### Min-Max Normalization
Scales to [0, 1].  
→ If used for training, must use training set min/max during inference.

### Standardization
Centers data (mean=0, std=1).  
→ For models sensitive to feature scale.

### Log Transformation
Normalizes **skewed distributions**.  
*Example: Income, web traffic.*

### One-Hot Encoding
Converts categorical to binary.  
*Example: Red, Blue → Red=1, Blue=0.*

### Label Encoding
Maps categories to integers.  
→ May introduce ordinal meaning unintentionally.

### Feature Binning
Buckets continuous features.  
*Example: Age groups — teen, adult, senior.*

### PCA (Principal Component Analysis)
Dimensionality reduction.  
→ Keeps key variance.

### Imputation
Handle missing values (mean, median, models).

### Outlier Handling
Remove or cap extreme values.

### Data Leakage
Ensure target info is NOT in input features during training.

---

## Built-In Algorithms

### K-Means
Unsupervised clustering.  
*Example: Customer segmentation.*

### PCA
Used before clustering or modeling to reduce noise.

### XGBoost
High-performance for tabular data.  
Supports classification, regression.

### LightGBM
Fast, efficient gradient boosting framework.  
Performs well on large, high-dimensional, and imbalanced datasets — supports techniques like class weights and custom loss functions.
*Example: Classification, ranking, regression.*  
Available in SageMaker with hyperparameter tuning.

### Linear Learner
Built-in for binary classification & regression.  
→ Fast and scalable.

### DeepAR
Time-series forecasting with probabilistic output.

### BlazingText
For text classification and word embeddings.

### Object2Vec
Learns embeddings for recommendations.

### Random Cut Forest
Anomaly detection.  
*Example: Fraud detection.*

### Seq2Seq
Sequence-to-sequence learning.  
*Example: Translation, summarization.*


---

## Hyperparameter Tuning

### Hyperband
Efficient search, stops bad configs early.  
→ Good for saving compute.

### Bayesian Optimization
Uses prior results to guide search.  
→ Best for fewer evaluations with complex models.

### Grid Search
Exhaustive search.  
→ Use when search space is small.

### Random Search
Faster than grid, samples randomly.  
→ Good for broad search quickly.

### Population-Based Training
Evolves hyperparameters during training.  
→ Common in deep learning.

### SageMaker Automatic Model Tuning
Built-in support for optimization using above strategies.

### Weight Decay
Regularization technique to prevent overfitting.  
Penalizes large weights in loss function.  
*Common in neural networks.*

---

## Troubleshooting / Problem Solving 

### Diagnosing Model Drift
- Use **SageMaker Model Monitor** to detect changes in input data distributions or prediction quality.
- Compare live prediction data to baseline (training) data.
- Retrain model with **recent data** to reflect the new distribution.
- Use **automated alerts** to monitor feature drift and data quality.
- Evaluate on a **fresh test set** to ensure updated model performs well.

### Reducing Training Time
- Increase **batch size** (fewer updates, more parallelism).
- Use **GPU**-backed instances (like `p3`, `g4`) for training.
- Implement **distributed training** (Horovod, SageMaker Distributed).
- Use **early stopping** when validation loss plateaus.
- Optimize **data pipeline**: Use **Pipe mode** or **RecordIO** for faster I/O.
- Profile training jobs with **SageMaker Debugger** to find bottlenecks.
- **Improve I/O performance** by moving S3 training data to **FSx for Lustre**:
  - FSx for Lustre automatically links with S3 and offers high-throughput, low-latency access to data.
- Automatically trigger training when new data is uploaded to S3:
  - Use EventBridge to kick off SageMaker Pipelines.

### Fixing Overfitting
Symptoms: High training accuracy, poor validation/test accuracy  
- Use **regularization** (`L1`, `L2`) to penalize complex models.
- Use **dropout layers** in neural networks.
- Simplify model (fewer parameters, shallower layers).
- Use **cross-validation** to better generalize across folds.
- Add **more training data** or apply **data augmentation**.
- Reduce **training epochs** to prevent overfitting.

### Fixing Underfitting
Symptoms: Low accuracy on both training and validation sets  
- Increase model complexity (more layers, deeper trees).
- **Train longer** — add more epochs.
- Improve **feature engineering** (new features, interactions).
- Reduce regularization if too restrictive.
- Switch to a **more powerful model** (e.g., from linear to non-linear).

### Handling Missing Data / Field Gaps
- Use **imputation** to fill missing values:  
  - **Mean/Median** for numerical fields  
  - **Mode** for categorical fields  
  - **Model-based** imputation for smarter filling (KNN, regression)  
- Drop rows or columns only if missing data is **minimal**.
- Mark missing values with an **indicator column** if meaningful.
- Ensure **same strategy** is applied at inference time (use pipeline or feature store).

### Deployment Types
- Use multi-model endpoints for serving many models efficiently from a single endpoint.
- Use serverless inference for spiky or infrequent traffic.
  - Limitations:
    - Hardware: CPU-only; does not support accelerated (GPU) instances.
    - Cold Starts: Can have cold start times if provisioned concurrency is not used.
    - Payload Size: Maximum of 6 MB.
    - Invocation Timeout: Maximum of 60 seconds.
- Use real-time inference for low-latency, high-throughput needs.
  - Unlike Serverless, it does support GPU and accelerated hardware.
  - Limitations:
    - Hardware: Supports both CPU and GPU instances.
    - Payload Size: Maximum of 6 MB.
    - Invocation Timeout: Default is 60 seconds; can be configured up to 15 minutes for large models.
- Use asynchronous inference for large payloads or long inference durations.
  - Limitations:
    - Payload Size: Supports payload sizes up to 1 GB.
    - Processing Time: Suitable for processing times up to 15 minutes.
- Use batch inference with for offline predictions on large datasets without deploying a real-time endpoint.
- Use model versioning and shadow testing for validating new models before replacing live ones.
- Use product variants to deploy multiple models (or versions) to the same endpoint and control traffic routing between them (e.g. 70/30 split for A/B testing).

### Debugging Model Performance
- Use **SageMaker Debugger** to inspect tensor values and training state.
- Analyze **confusion matrix**, **precision/recall per class**, and **feature importance**.
- Use **Model Explainability** (SHAP) in SageMaker Clarify to interpret predictions.
- Monitor input **data quality** — missing values, outliers, or drift can degrade performance.

### Reducing Model Size (for Deployment Efficiency)
- Use **SageMaker Neo** to compile and optimize models for specific hardware targets (CPU, GPU, edge).
  - Improves latency and reduces memory usage.
  - Supports models from TensorFlow, PyTorch, XGBoost, etc.
- Apply **Quantization** to reduce weights from float32 to int8 or float16.
  - Smaller size, faster inference with minimal accuracy loss.
- Use **Pruning** to remove unimportant connections or weights in the model.
  - Speeds up inference and reduces model footprint.
- Compress or reduce **input features**:
  - Drop low-importance features (based on SHAP, feature importance).
  - Apply dimensionality reduction (e.g., **PCA**) before training.
  - Helps with latency and inference speed on constrained devices.

### Reducing ML Costs
- **Use Spot Instances** for training to reduce EC2 cost by up to 90%.  
  → Use **Managed Spot Training** in SageMaker to auto-recover from interruptions.
- Use **SageMaker Savings Plans** to commit to ML instance usage and reduce long-term costs.
  → Applies to training and inference on SageMaker for flexible instance families and regions.
- Use **SageMaker Serverless Inference** for low-volume, spiky traffic to avoid idle instance cost.
- Use **SageMaker Multi-Model Endpoints** to host multiple models on a single endpoint.
- Use **AWS Computer Optimizer.**
- Use **AWS Budgets** to set cost or usage alerts and catch unexpected ML spend early.
- Use **SageMaker Batch Transform** for batch jobs instead of provisioning real-time endpoints.
- Use AWS Batch with Spot Instances for long-running or GPU-based containerized ML jobs.
- Reduce **instance size or type** (e.g., use `ml.m5` instead of `ml.c5` for CPU-bound tasks).
- **Use SageMaker Debugger** and **Profiler** to optimize resource usage during training.
- Compress models using **quantization** or **pruning** to reduce inference costs.
- Offload preprocessing to **AWS Glue**, **Lambda**, or **Athena** to minimize compute use in training.

### Securing ML Workloads
- Use **inter-node encryption** during training to secure data in transit.
- Run jobs in **network isolation** mode to disable internet access entirely.
- Deploy SageMaker training and inference within a **private VPC subnet** for full network control.
- Use **VPC endpoints to S3** to ensure data stays within AWS's network.
- Apply **IAM roles with least privilege** for access to training data and model artifacts.
- Use **KMS** to encrypt training data, model outputs, EBS volumes, and logs.
- Store API keys or secrets in **AWS Secrets Manager**, which supports rotation.
- Use **AWS Lake Formation** to define and enforce fine-grained access controls and data governance policies on S3 data used in ML workflows.
- Enable **AWS CloudTrail** to log and audit all API activity related to SageMaker, IAM, S3, and KMS — helps with compliance and security investigations.
- Use **Amazon Macie** or **Amazon Comprehend** to automatically detect and classify PII in S3 and text data.
- Use **AWS Glue DataBrew** to anonymize, mask, or redact PII data before using it in ML workflows.

### Improving Latency
- **ModelLoadingWaitTime** indicates how long it takes to load the model container on the endpoint — high values = cold start latency.
- Use **Provisioned Inference Endpoints** for consistently low latency — keeps containers pre-loaded and ready.
- Use **Real-Time Inference Endpoints** for general low-latency needs with auto-scaling — but expect cold starts during initial invocation or scaling events.
- For latency-sensitive production workloads, prefer **Provisioned Inference** over Real-Time Inference to avoid startup delay.
- Enable **Warm Pools** to reduce container loading time when scaling Real-Time endpoints.
- Set **buffering to 0** in Firehose/Kinesis to reduce delays in delivering streaming data to endpoints.
- Keep **training and inference data in the same Region and Availability Zone** to minimize data transfer latency.
- Use real-time services (like Kinesis, Flink, Lambda, Amazon MSK, SageMaker Real-Time Inference) for fast, low-latency processing — avoid batch-oriented tools like Glue, Athena, QuickSight, SageMaker Batch Transform, or EMR for latency-critical pipelines.

### Reducing Model Bias
- Use **SageMaker Clarify** to detect:
  - **Pre-training bias** (e.g., biased input data)
  - **Post-training bias** (e.g., biased predictions)
  - **Feature importance** (identify if a sensitive attribute influences outcomes)
- Ensure training data is **balanced** across sensitive groups (e.g., gender, age).
- Avoid using **proxy features** (features that correlate with sensitive attributes).
- Use **differential fairness** metrics (like disparate impact) to quantify bias.
- Retrain with **balanced or reweighted samples** to reduce bias.

### Too Many False Positives in Fraud Detection
- Review **precision-recall tradeoff**:
  - **Increase precision** by adjusting classification threshold.
- Use **class weights or cost-sensitive learning** if fraud class is underrepresented.
- Add **more representative fraud samples** or perform **SMOTE/oversampling**.
- Consider **ensemble methods** to reduce overfitting and improve generalization.
- Use **domain-specific rules** in conjunction with ML output for hybrid filtering.

### Model Training Oscillates / Does Not Converge
- Symptoms: Loss fluctuates wildly or does not reduce consistently.
- Possible fixes:
  - **Lower the learning rate** — too high a rate can cause overshooting.
  - **Use learning rate decay** or schedulers.
  - Check for **data issues** (e.g., outliers, label noise).
  - Ensure **shuffling of training data** to prevent learning in batches.
  - Use **gradient clipping** in deep learning models.
  - Use **batch normalization** to stabilize training.
  - Try a more stable optimizer like **Adam** instead of SGD.

### Model Training Gets Stuck in Local Minimum
- Training seems stable but the model converges to a **suboptimal solution**.
- Possible causes:
  - Poor initialization of weights
  - Inappropriate learning rate
  - Limited model complexity
- Fixes:
  - Use **different/random weight initializations**
  - Add **momentum** to gradient descent
  - Increase model complexity if underfitting
  - Try **learning rate scheduling** to escape flat regions or local minima

### Senstive Information in Data 
- Mask it out: use AWS Glue or DataBrew to apply data masking or obfuscation
- Detect it (PII): use AWS Macie (for S3) or Comprehend (NLP-based)
- For Redshift: use Dynamic Data Masking — hides data at query time without changing source data

### Duplicate Data in Training Set
- Use **AWS Glue FindMatches transform** to detect fuzzy duplicates in records.
- Useful when duplicates are not exact matches (e.g., different casing, typos).

### Anomalies in Training Data
- Use **SageMaker Data Wrangler** → “**Data Quality and Insights Report**”.
- Helps find outliers, missing values, duplicates, and skewed distributions.

### Class Imbalance
- Oversample the minority class using:
  - SMOTE or random oversampling techniques.
  - Assign class weights to reduce bias toward majority class.


---

## AWS ML Infrastructure

### AWS Trainium
Purpose-built accelerator for **deep learning training**.  
- Optimized for high throughput, scalability, and low cost.  
- Supported via **Trn1 instances** in SageMaker and EC2.  
- Best for **large-scale transformer training**, foundation models.

### AWS Inferentia
Custom chip for **inference workloads**.  
- Use with **Inf1 or Inf2 instances** on SageMaker.  
- Optimized for low latency and cost-efficient inference at scale.  
- Best for **production deployment of deep learning models**.

### AWS Deep Learning Containers (DLCs)
Pre-built Docker containers optimized for ML workloads.  
- Includes popular frameworks: TensorFlow, PyTorch, MXNet, HuggingFace, etc.  
- Fully compatible with **SageMaker**, **ECS**, **EKS**, and **EC2**.  
- Reduces setup time and ensures **optimized GPU/CPU performance**.  
- Ideal for **custom training**, experimenting locally, or deploying in hybrid environments.

### EC2 Instances for ML
High-performance compute options for training or inference jobs.

#### GPU Instances:
- **p3**: For **training deep learning models** (e.g., neural networks, transformers).
- **p4d**: Newest instance with **NVIDIA A100** GPUs for **high-performance training**.
- **g4ad, g4dn**: For **ML inference** and **medium-scale model training**.

#### CPU Instances:
- **m5, c5**: General-purpose or compute-optimized EC2 instances for less intensive ML tasks.

### AWS Elastic Inference
Attaches **GPU acceleration** to **CPU instances** for inference workloads.  
- Cost-effective alternative to full GPU instances.
- Supports **TensorFlow**, **PyTorch**, **MXNet**, and **ONNX**.

### AWS Lambda for Serverless Inference
- Run inference **serverless** without provisioning infrastructure.  
- Best for **low-volume, low-latency inference**.
- Integrates with **SageMaker** and can handle models deployed on **SageMaker endpoints**.

### AWS Elastic Kubernetes Service (EKS)
Managed Kubernetes service that simplifies **containerized model deployment**.  
- Use **Kubernetes operators** to manage **SageMaker endpoints** in a containerized environment.
- Ideal for **large-scale ML deployments** and **scalable ML workflows**.

### Amazon Elastic Container Service (ECS)
Managed container orchestration service for **large-scale inference** workloads.  
- Can be used with **Deep Learning Containers** for scalable **inference and training**.
- Supports integration with **SageMaker** for end-to-end ML workflows.

### AWS Batch
Fully managed batch processing service for running large-scale ML jobs.  
- Ideal for jobs requiring **massive parallelization** (e.g., hyperparameter tuning).
- Supports **SageMaker Training Jobs**, Docker containers, and **GPU instances**.

### AWS Outposts
Deploys **AWS infrastructure on-premises** for **hybrid ML workloads**.  
- Useful for compliance-sensitive industries or applications with **low-latency requirements**.
- Integrates seamlessly with **SageMaker** and **other AWS services**.

### AWS Snowball Edge
Data transfer device with **edge computing capabilities** for ML tasks in remote environments.  
- Perform **local inference** and **data processing** when low-latency is critical and connectivity is limited.
- Can be used to **train models** or infer locally, and then transfer data to the cloud for further analysis.

---

## Top 40 General Tips for MLA-C01 (Ordered by Importance)

1. **Use Managed AWS Services**: Always default to managed options like SageMaker, Glue, Bedrock unless you need full control.

2. **Minimize Operational Overhead**: Pick serverless (Lambda, Firehose, Glue) when possible for low-maintenance solutions.

3. **SageMaker Model Monitor = Detect Drift**: Use it to track feature distribution, data quality, and prediction drift.

4. **SageMaker Clarify = Detect Bias**: Use it pre- and post-deployment to ensure fairness and explainability.

5. **Hyperparameter Tuning = Use SageMaker AMT**: Automatic Model Tuning is the go-to for optimizing training jobs.

6. **Real-Time Inference = SageMaker Endpoints**: Use real-time endpoints for synchronous, low-latency predictions.

7. **Batch Inference = SageMaker Batch Transform**: Ideal for large volumes of data processed asynchronously.

8. **Async Inference = SageMaker Async Endpoints**: For large or slow inference requests with predictable duration.

9. **SageMaker Pipelines = Full MLOps Automation**: Preferred tool for building, training, evaluating, and deploying models.

10. **Use Spot Instances for Training**: Cuts cost significantly for non-critical or retryable jobs.

11. **Multi-AZ = High Availability**: Choose multi-AZ designs for fault tolerance in production workloads.

12. **Multi-Region = Disaster Recovery**: Use for business continuity and compliance across geographies.

13. **Minimize Overfitting = Use Regularization**: Techniques like L1, L2, dropout, and early stopping reduce model overfitting.

14. **Prefer Foundation Models for NLP = Use Bedrock**: For GenAI or NLP tasks, leverage pre-trained models via Bedrock.

15. **Low-Frequency Inference = Use Serverless Endpoints**: Cost-effective for occasional inference workloads.

16. **Model Reuse = SageMaker Feature Store**: Store and reuse features across multiple training jobs.

17. **CI/CD = CodePipeline + SageMaker or Step Functions**: Automate ML workflows from source to deploy.

18. **Edge Deployment = Use SageMaker Neo**: Compile and optimize models for deployment on edge devices.

19. **Secure Access = Use IAM with Least Privilege**: Always restrict roles and policies to just what's needed.

20. **Secure Endpoints = Deploy in VPC**: Isolate endpoints and training jobs inside a secure VPC.

21. **Scaling = Use Auto Scaling for Endpoints**: Configure invocations per instance or latency-based scaling.

22. **Monitor Everything = CloudWatch Logs, Metrics, Alarms**: Essential for debugging and tracking behavior.

23. **Audit Everything = Use CloudTrail**: Log all actions and events for security and traceability.

24. **Right-Size Inference = Use Inference Recommender**: Helps choose the best instance type based on latency and cost.

25. **Right-Size Training = Use Compute Optimizer**: Optimize training instances using usage data.

26. **Load Testing = Use Shadow Deployments**: Compare new models to production without customer impact.

27. **Trigger Workflows = Use EventBridge**: Automate pipelines or retraining when upstream events occur.

28. **Track Experiments = Use SageMaker Experiments**: Manage and compare model training runs and metadata.

29. **Secure CI/CD = IAM + Pipeline Encryption**: Apply access controls and encryption to protect pipelines.

30. **Join S3 + Redshift = Use Redshift Spectrum**: Query S3 directly using Redshift SQL.

31. **Low Latency Streaming = Use Kinesis Streams**: Best for sub-second processing needs.

32. **Near Real-Time = Use Kinesis Firehose**: Easier ingestion to S3/Redshift with minimal config.

33. **Fast File Transfer = Use S3 Transfer Acceleration**: Optimize uploads/downloads for global teams.

34. **Automate Infrastructure = Use CloudFormation or CDK**: Ensure reproducibility and CI/CD of infrastructure.

35. **No-Code ETL = Use DataBrew**: Quick transformations for non-engineers or prototyping.

36. **Rule-Based Data Checks = Use Glue Data Quality**: Validate schema, nulls, custom thresholds.

37. **Split Data Properly = Shuffle + Stratify**: Prevent data leakage and ensure balanced train/test sets.

38. **Explainability = Use SHAP with Clarify**: Interpret feature influence for compliance or trust.

39. **Anomaly Detection = Use Lookout for Metrics**: Automatically flag metric anomalies in production.

40. **Labeling = Use Ground Truth or Mechanical Turk**: For building high-quality labeled datasets at scale.



## Domain 1: Data Preparation for Machine Learning (ML)

---

### Task 1.1: Ingest and Store Data

#### Knowledge Of:
- Data Formats: Parquet, JSON, CSV, ORC, Avro, RecordIO  
  - Parquet/ORC = efficient columnar formats → best for analytics  
  - RecordIO = used in SageMaker training

- AWS Storage Services:
  - Amazon S3: Object store, scalable, cost-effective  
  - Amazon EFS: Scalable file system for Linux-based workloads  
  - FSx for NetApp ONTAP: Shared storage for enterprise apps

- Streaming Sources:
  - Kinesis Data Streams / Firehose  
  - Amazon MSK (Managed Streaming for Kafka)  
  - Amazon Flink (via Kinesis Data Analytics) for real-time processing

#### Skills In:
- Extracting Data from AWS Sources:
  - Use S3 Transfer Acceleration for faster global uploads
  - Use EBS Provisioned IOPS for high-performance access

- Choosing Formats by Access Pattern:
  - Columnar (Parquet, ORC) = read-heavy, analytics  
  - Row-based (JSON, CSV) = write-heavy, logs

- SageMaker Integrations:
  - Load data into SageMaker Data Wrangler  
  - Store features in SageMaker Feature Store

- Merging Data:
  - Use AWS Glue, Spark on EMR, or pandas

- Troubleshooting:
  - Capacity: check storage throughput limits (e.g., EBS)
  - Scalability: S3 scales automatically, EFS throughput grows with usage

- Storage Decisions:
  - S3: Default, cost-effective
  - EBS: Block storage, high performance
  - EFS/FSx: Shared file systems for ML training

**Rule of Thumb**  
Parquet for Analytics: Choose Parquet or ORC for ML workflows needing efficient analytics and compression.  
S3 for Raw Data: Use S3 as the default storage layer unless latency/performance requires otherwise.

---

### Task 1.2: Transform Data and Perform Feature Engineering

#### Knowledge Of:
- Data Cleaning Techniques:
  - Detecting outliers, imputing missing values, deduplication, combining data

- Feature Engineering:
  - Scaling, standardization, binning, log transforms, normalization

- Encoding:
  - One-hot encoding, label encoding, binary encoding, tokenization

- Tools:
  - SageMaker Data Wrangler
  - AWS Glue and Glue DataBrew (code/no-code)
  - AWS Lambda, Spark on EMR for stream transformation

- Labeling Services:
  - SageMaker Ground Truth
  - Amazon Mechanical Turk

#### Skills In:
- Transforming Data:
  - Glue for scalable ETL  
  - Data Wrangler for fast exploration  
  - Spark on EMR for big data

- Creating Features:
  - Use SageMaker Feature Store to register and reuse features

- Labeling:
  - Use SageMaker Ground Truth with built-in workflows

**Rule of Thumb**  
Data Wrangler for Rapid Prototyping: Use for visual data prep and export to SageMaker pipelines.  
Glue for Scalable ETL: Best for large-scale or scheduled data transformations.

---

### Task 1.3: Ensure Data Integrity and Prepare Data for Modeling

#### Knowledge Of:
- Bias Metrics:
  - Class Imbalance (CI)
  - Difference in Proportions of Labels (DPL)

- Bias Mitigation Strategies:
  - Resampling, Synthetic data generation
  - Use SageMaker Clarify for bias detection

- Data Security:
  - Encryption: S3 SSE, KMS
  - Masking, anonymization

- Compliance Considerations:
  - PII, PHI, data residency

#### Skills In:
- Data Quality Validation:
  - AWS Glue Data Quality
  - AWS Glue DataBrew

- Bias Detection and Mitigation:
  - Use SageMaker Clarify
  - Apply data augmentation, shuffling, class balancing

- Preparing for Training:
  - Use Amazon EFS or FSx as input sources
  - Ensure correct format and folder structure for training jobs

**Rule of Thumb**  
Clarify for Bias: Use SageMaker Clarify to measure and mitigate pre-training bias.  
Data Quality: Use Glue Data Quality or DataBrew for validating schema, completeness, and integrity.

---

## Domain 2: ML Model Development

---

### Task 2.1: Choose a Modeling Approach

#### Knowledge Of:
- ML Algorithm Capabilities:
  - Classification, regression, clustering, time series forecasting
  - Trade-offs: performance, interpretability, complexity, cost

- AWS AI Services:
  - Amazon Translate: Text translation
  - Amazon Transcribe: Speech-to-text
  - Amazon Rekognition: Image/video analysis
  - Amazon Bedrock: Foundation model (FM) access (Claude, Titan, etc.)

- Interpretability:
  - Consider explainability when choosing models (e.g., avoid deep nets for regulated environments)

- SageMaker Built-in Algorithms:
  - XGBoost, Linear Learner, KNN, Factorization Machines, BlazingText
  - Best for rapid prototyping with minimal code

#### Skills In:
- Assess Feasibility:
  - Use business constraints and data availability to assess model viability

- Choose Appropriate Models:
  - Classification → Logistic regression, XGBoost
  - Text Gen → Foundation Models in Bedrock
  - Image → CNNs or Rekognition

- Use Pre-built Tools:
  - JumpStart: Solution templates with pretrained models
  - Bedrock: Use if pre-trained FM can meet needs faster and cheaper

- Optimize for Cost:
  - Lightweight models for inference on edge devices
  - Prefer AI services over building custom models if the task is common

**Rule of Thumb**  
Use AI Services (Translate, Transcribe, Rekognition) when you don't need custom training.  
Use Bedrock or JumpStart to quickly prototype with Foundation Models or pre-trained templates.

---

### Task 2.2: Train and Refine Models

#### Knowledge Of:
- Training Concepts:
  - Epochs, batch size, steps affect training time and accuracy

- Reduce Training Time:
  - Use early stopping, distributed training, mixed precision

- Improve Model Performance:
  - Hyperparameter tuning, feature engineering, model selection
  - Regularization (dropout, L1/L2, weight decay) to avoid overfitting

- Hyperparameter Tuning:
  - SageMaker AMT supports random search, Bayesian optimization

- External Model Integration:
  - Import models via script mode or model.tar.gz for custom frameworks

#### Skills In:
- Use SageMaker Built-in Algorithms or ML Frameworks (TensorFlow, PyTorch)
- Fine-Tune Pre-trained Models using custom data (JumpStart, Bedrock)
- Apply Hyperparameter Tuning with SageMaker AMT
- Prevent Overfitting:
  - Use regularization, early stopping, dropout
- Model Ensembling:
  - Combine models (bagging, boosting, stacking) for better performance
- Reduce Model Size:
  - Quantization, pruning, compressing inputs/features
- Version Management:
  - Use SageMaker Model Registry to track versions and enable audits

**Rule of Thumb**  
Use SageMaker AMT for hands-off hyperparameter tuning.  
Reduce overfitting with dropout and regularization, underfitting with deeper models or more features.

---

### Task 2.3: Analyze Model Performance

#### Knowledge Of:
- Evaluation Metrics:
  - Classification: Accuracy, Precision, Recall, F1 Score, AUC-ROC
  - Regression: RMSE, MAE
  - Visualization: Confusion matrix, ROC curves, heat maps

- Baselines:
  - Use null models or historical data for baseline comparison

- Overfitting/Underfitting Detection:
  - High train/low test accuracy = overfit  
  - Low train accuracy = underfit

- Interpretability with Clarify:
  - SHAP values, bias metrics

- Convergence Issues:
  - Use SageMaker Debugger to identify stalled or unstable training

#### Skills In:
- Select Metrics Based on Problem Type:
  - Imbalanced classes → use F1, precision/recall
  - Continuous output → use RMSE, MAE

- Evaluate Trade-offs:
  - Training time vs performance vs cost

- Reproducibility:
  - Use SageMaker Experiments for tracking runs

- Shadow Testing:
  - Compare production and shadow variants (A/B testing)

- Debug Models:
  - Use SageMaker Debugger to inspect gradients, loss functions

**Rule of Thumb**  
Use F1 or AUC-ROC for imbalanced classification problems.  
Use SageMaker Clarify for model bias and interpretability; Debugger for training issues.

---

## Domain 3: Deployment and Orchestration of ML Workflows

---

### Task 3.1: Select Deployment Infrastructure Based on Existing Architecture and Requirements

#### Knowledge Of:
- Deployment Best Practices:
  - Versioning, rollback strategies, blue/green, canary deployments

- AWS Deployment Services:
  - SageMaker: Supports real-time, async, and batch inference endpoints

- Serving Models:
  - Real-time: Low latency, synchronous (use SageMaker real-time endpoints)
  - Async: For long-running jobs (SageMaker async endpoints)
  - Batch: For large batch scoring (SageMaker batch transform)

- Compute Provisioning:
  - Choose GPU for deep learning models, CPU for lightweight inference
  - Match instance family to model size and inference latency needs

- Endpoint Types:
  - Real-time endpoints
  - Asynchronous endpoints
  - Batch transform
  - Serverless endpoints for infrequent low-latency needs

- Containers:
  - Use SageMaker pre-built containers or bring your own (BYOC)

- Edge Optimization:
  - Use SageMaker Neo to compile and optimize models for edge deployment

#### Skills In:
- Evaluate Trade-offs:
  - Cost vs latency vs throughput
  - Spot vs On-Demand

- Choose Deployment Targets:
  - SageMaker endpoints, Lambda (for light models), ECS/EKS for containerized workflows

- Orchestration Tools:
  - SageMaker Pipelines: ML-specific CI/CD  
  - Apache Airflow: General workflow orchestration (via MWAA)

- Multi-model Hosting:
  - Serve multiple models behind a single endpoint

**Rule of Thumb**  
Use SageMaker real-time endpoints for low-latency inference, async for large/predictable delay tasks, batch for offline scoring.  
Use SageMaker Neo when deploying models to edge devices.

---

### Task 3.2: Create and Script Infrastructure Based on Existing Architecture and Requirements

#### Knowledge Of:
- Resource Types:
  - On-demand: Pay-per-use, no upfront commitment  
  - Provisioned: Reserved capacity with guaranteed availability

- Scaling Policies:
  - Automatic scaling (invocations per instance, CPU utilization, latency)
  - Manual and scheduled scaling

- Infrastructure as Code (IaC):
  - CloudFormation: Declarative templates  
  - AWS CDK: Imperative IaC with familiar languages

- Containerization:
  - Use ECR to store images  
  - ECS/EKS for container deployment  
  - BYOC to bring custom environments to SageMaker

- SageMaker Auto Scaling:
  - Use endpoint invocation metrics to auto-scale based on traffic or time

#### Skills In:
- Enable Cost-effective Scaling:
  - Use Spot Instances for training  
  - Use Lambda + SageMaker endpoint for low-frequency workloads

- Automate Provisioning:
  - CloudFormation/CDK to deploy entire ML infrastructure stacks

- Container Management:
  - Build containers, push to ECR, deploy with ECS/EKS or SageMaker

- VPC Configurations:
  - Configure SageMaker endpoints to run inside a VPC for security

- SageMaker SDK:
  - Use SDK to deploy, update, and manage model endpoints

**Rule of Thumb**  
Use auto-scaling for inference endpoints to balance cost and performance.  
Use CloudFormation/CDK for reproducible and automated infrastructure deployments.

---

### Task 3.3: Use Automated Orchestration Tools to Set Up CI/CD Pipelines

#### Knowledge Of:
- Code Tools:
  - CodePipeline: CI/CD orchestration  
  - CodeBuild: Build & test stages  
  - CodeDeploy: Deployment stages

- CI/CD in ML:
  - Includes model retraining, testing, validation, and deployment

- Deployment Strategies:
  - Canary, blue/green, and linear rollouts for safe updates

- Code Repositories:
  - Git/GitHub as version control and CI/CD triggers

#### Skills In:
- Pipeline Configuration:
  - Define stages for training, evaluation, deployment

- Git-Based Triggers:
  - Use GitHub Flow or Gitflow for CI/CD pipeline branching and release

- Automate Workflows:
  - Use SageMaker Pipelines or EventBridge to trigger retraining or inference

- Testing:
  - Unit, integration, and end-to-end tests in model deployment pipelines

- Retraining Triggers:
  - Build pipelines that automatically retrain models when data or code changes

**Rule of Thumb**  
Use SageMaker Pipelines for ML-specific CI/CD with built-in tracking.  
Use CodePipeline + CodeBuild + CodeDeploy for general purpose CI/CD and integration with Git.

---

## Domain 4: ML Solution Monitoring, Maintenance, and Security

---

### Task 4.1: Monitor Model Inference

#### Knowledge Of:
- Model Drift:
  - Concept drift: relationship between input and output changes
  - Data drift: input distribution changes over time

- Monitoring Techniques:
  - Track data quality, prediction distributions, feature skew

- ML Lens Principles:
  - Design for observability, detectability, and traceability in ML workflows

#### Skills In:
- Use SageMaker Model Monitor:
  - Monitor input data, prediction distributions, feature attribution drift

- Detect Anomalies:
  - Set alerts for data pipeline errors or missing features

- Monitor with Clarify:
  - Analyze post-deployment bias and data drift

- A/B Testing:
  - Deploy shadow models to monitor performance before full rollout

**Rule of Thumb**  
Use SageMaker Model Monitor for drift detection and prediction monitoring.  
Use Clarify for bias and drift analysis during and after deployment.

---

### Task 4.2: Monitor and Optimize Infrastructure and Costs

#### Knowledge Of:
- ML Infra Metrics:
  - Utilization, latency, throughput, fault tolerance

- Monitoring Tools:
  - AWS X-Ray: Trace app performance  
  - CloudWatch Logs & Insights: Log aggregation and queries  
  - Lambda Insights: Function-specific monitoring

- Logging and Audit:
  - Use AWS CloudTrail for tracking resource usage and security events

- Instance Types:
  - Memory optimized: R-series  
  - Compute optimized: C-series  
  - General purpose: T-series  
  - Inference optimized: Inf1, G5

- Cost Management Tools:
  - Cost Explorer, Budgets, Trusted Advisor, Billing & Cost Management

#### Skills In:
- Use CloudWatch Dashboards:
  - Visualize performance metrics

- Configure Logs and Alarms:
  - Trigger alerts based on latency, throughput, or cost spikes

- CloudTrail Integration:
  - Log activities and invoke retraining via EventBridge

- Instance Right-Sizing:
  - Use SageMaker Inference Recommender  
  - Use Compute Optimizer for instance recommendations

- Cost Optimization:
  - Spot for training, Reserved for predictable workloads  
  - Apply tagging strategy for allocation

**Rule of Thumb**  
Use CloudWatch for monitoring, CloudTrail for auditing, and Cost Explorer for financial analysis.  
Use Spot for training, and Reserved or Savings Plans for steady workloads.

---

### Task 4.3: Secure AWS Resources

#### Knowledge Of:
- IAM Roles and Policies:
  - Use least privilege principles  
  - Role-based access for SageMaker, S3, ECR, etc.

- SageMaker Security Features:
  - SageMaker Role Manager for access setup  
  - Model encryption with KMS  
  - Endpoint VPC isolation

- Network Controls:
  - Use VPC, subnets, security groups to isolate ML workloads

- Secure CI/CD:
  - Secure CodePipeline stages with IAM, encryption, and audits

#### Skills In:
- Configure IAM:
  - Role assumption, trust relationships, policy scoping

- Audit and Monitor:
  - Use CloudTrail for compliance  
  - Set up CloudWatch and EventBridge for monitoring anomalies

- Secure Networking:
  - Deploy endpoints within VPC  
  - Restrict access with security groups and NACLs

- Debug Security:
  - Trace IAM permission errors and troubleshoot network isolation issues

**Rule of Thumb**  
Always use least privilege IAM policies and isolate endpoints in VPCs.  
Enable encryption and logging to maintain compliance and traceability.

---

## Appendix: In-Scope AWS Services for MLA-C01

---

### Analytics

- **Amazon Athena**: Serverless SQL over S3 data.
- **Amazon Data Firehose**: Stream data into S3/Redshift with optional transform.
- **Amazon EMR**: Managed big data processing (Spark, Hadoop).
- **AWS Glue**: Serverless ETL for large-scale data prep.
- **AWS Glue DataBrew**: No-code data wrangling.
- **AWS Glue Data Quality**: Rule-based data validation and profiling.
- **Amazon Kinesis**: Real-time data streaming.
- **AWS Lake Formation**: Secure and govern S3 data lakes.
- **Amazon Managed Service for Apache Flink**: Real-time stream processing.
- **Amazon OpenSearch Service**: Search and analytics engine.
- **Amazon QuickSight**: BI and dashboards over AWS data.
- **Amazon Redshift**: Scalable data warehouse with ML integration.

---

### Application Integration

- **Amazon EventBridge**: Event bus for triggering workflows.
- **Amazon MWAA**: Managed Apache Airflow for orchestration.
- **Amazon SNS**: Pub/sub messaging for decoupled systems.
- **Amazon SQS**: Message queue for decoupling components.
- **AWS Step Functions**: Visual workflow orchestration.

---

### Cloud Financial Management

- **AWS Billing and Cost Management**: View and manage charges.
- **AWS Budgets**: Set cost and usage alerts.
- **AWS Cost Explorer**: Visualize usage and spending trends.

---

### Compute

- **AWS Batch**: Run batch jobs on managed compute.
- **Amazon EC2**: Scalable virtual machines.
- **AWS Lambda**: Serverless compute for event-driven ML.
- **AWS Serverless Application Repository**: Deploy Lambda-based blueprints.

---

### Containers

- **Amazon ECR**: Container image registry.
- **Amazon ECS**: Managed container orchestration.
- **Amazon EKS**: Managed Kubernetes service.

---

### Database

- **Amazon DocumentDB**: Managed MongoDB-compatible DB.
- **Amazon DynamoDB**: Serverless key-value and document DB.
- **Amazon ElastiCache**: In-memory caching (Redis/Memcached).
- **Amazon Neptune**: Graph database.
- **Amazon RDS**: Managed relational databases (MySQL, Postgres, etc.).

---

### Developer Tools

- **AWS CDK**: Code-based IaC using TypeScript, Python, etc.
- **AWS CodeArtifact**: Package management for code dependencies.
- **AWS CodeBuild**: Build and test automation.
- **AWS CodeDeploy**: Automate deployment to compute targets.
- **AWS CodePipeline**: CI/CD pipeline service.
- **AWS X-Ray**: Trace application and ML pipeline latency.

---

### Machine Learning

- **Amazon A2I**: Human-in-the-loop review for ML predictions.
- **Amazon Bedrock**: Access to foundation models via API.
- **Amazon CodeGuru**: ML-based code reviews and profiling.
- **Amazon Comprehend**: NLP for sentiment, entity detection.
- **Amazon Comprehend Medical**: NLP for healthcare data.
- **Amazon DevOps Guru**: Operational ML anomaly detection.
- **Amazon Fraud Detector**: Build fraud detection models.
- **AWS HealthLake**: Store and analyze healthcare data (FHIR).
- **Amazon Kendra**: Intelligent search engine.
- **Amazon Lex**: Conversational chatbot engine.
- **Amazon Lookout for Equipment**: Predictive maintenance for IoT.
- **Amazon Lookout for Metrics**: Anomaly detection in metrics.
- **Amazon Lookout for Vision**: Visual anomaly detection.
- **Amazon Mechanical Turk**: Human labeling workforce.
- **Amazon Personalize**: Recommendation engine.
- **Amazon Polly**: Text-to-speech.
- **Amazon Q**: Generative AI-powered assistant.
- **Amazon Rekognition**: Image/video analysis.
- **Amazon SageMaker**: Full ML lifecycle platform.
- **Amazon Textract**: Extract text/structured data from docs.
- **Amazon Transcribe**: Speech-to-text.
- **Amazon Translate**: Language translation.

---

### Management and Governance

- **AWS Auto Scaling**: Scale resources dynamically.
- **AWS Chatbot**: Alerts and notifications in Slack/Chime.
- **AWS CloudFormation**: Declarative infrastructure as code.
- **AWS CloudTrail**: Audit logs for actions on AWS.
- **Amazon CloudWatch**: Metrics, logging, dashboards.
- **Amazon CloudWatch Logs**: Store and analyze logs.
- **AWS Compute Optimizer**: Instance size/cost recommendations.
- **AWS Config**: Track configuration changes.
- **AWS Organizations**: Manage multiple AWS accounts.
- **AWS Service Catalog**: Pre-approved infrastructure templates.
- **AWS Systems Manager**: Operational insights and automation.
- **AWS Trusted Advisor**: Best practice and cost checks.

---

### Media

- **Amazon Kinesis Video Streams**: Ingest and process video streams.

---

### Migration and Transfer

- **AWS DataSync**: Accelerate on-prem to cloud data transfers.

---

### Networking and Content Delivery

- **Amazon API Gateway**: Create and manage APIs.
- **Amazon CloudFront**: CDN for content acceleration.
- **AWS Direct Connect**: Private network link to AWS.
- **Amazon VPC**: Isolated network for resources.

---

### Security, Identity, and Compliance

- **AWS IAM**: Identity and access management.
- **AWS KMS**: Encryption key management.
- **Amazon Macie**: Managed PII detection in S3.
- **AWS Secrets Manager**: Store and rotate credentials.

---

### Storage

- **Amazon EBS**: Block storage for EC2.
- **Amazon EFS**: Scalable shared file storage.
- **Amazon FSx**: Managed Windows/NetApp file systems.
- **Amazon S3**: Scalable object storage.
- **Amazon S3 Glacier**: Archival storage.
- **AWS Storage Gateway**: Hybrid storage integration.

---
